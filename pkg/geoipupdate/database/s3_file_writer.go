package database

import (
	"crypto/md5"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/gofrs/flock"
	"github.com/pkg/errors"
	"hash"
	"io"
	"log"
	"os"
	"strings"
	"time"
)

const DateModifiedTag = "DateOfSourceDatabaseModification"
const Encryption = "AES256"

//S3DatabaseWriter is a databaseWriter that stores the database to a target s3 bucket and key
type S3DatabaseWriter struct {
	s3Client          *s3.S3
	s3Bucket          string
	s3Key             string
	disableEncryption bool
	lock              *flock.Flock
	oldHash           string
	multiWriter       io.Writer
	temporaryFile     *os.File
	md5Writer         hash.Hash
}

//NewS3DatabaseWriter creates a new S3DatabaseWriter, creating necessary locks and temporary files to protect from
//	concurrent writes
func NewS3DatabaseWriter(s3Client *s3.S3, s3Bucket, s3Key, lockFile string, verbose bool) (*S3DatabaseWriter, error) {
	dbWriter := &S3DatabaseWriter{
		s3Client:          s3Client,
		s3Bucket:          s3Bucket,
		s3Key:             s3Key,
		disableEncryption: false,
	}

	var err error
	if err = dbWriter.retrieveOldMD5Hash(); err != nil {
		return nil, err
	}
	if dbWriter.lock, err = CreateLockFile(os.TempDir(), lockFile, verbose); err != nil {
		return nil, err
	}

	keyPath := strings.Split(s3Key, "/")
	temporaryFilename := fmt.Sprintf("%s/%s.temporary", os.TempDir(), keyPath[len(keyPath)-1])
	dbWriter.temporaryFile, err = os.OpenFile(temporaryFilename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return nil, errors.Wrap(err, "error creating temporary file")
	}
	dbWriter.md5Writer = md5.New()
	dbWriter.multiWriter = io.MultiWriter(dbWriter.md5Writer, dbWriter.temporaryFile)

	return dbWriter, nil
}

//retrieveOldMD5Hash uses the s3 bucket and key to query for the ETag (the MD5) for the S3 object
func (writer *S3DatabaseWriter) retrieveOldMD5Hash() error {
	response, err := writer.s3Client.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(writer.s3Bucket),
		Key:    aws.String(writer.s3Key),
	})
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case s3.ErrCodeNoSuchKey:
				writer.oldHash = zeroMD5
				return nil
			default:
				return errors.Wrapf(err, "error fetching object %s/%s", writer.s3Bucket, writer.s3Key)
			}
		}
		return errors.Wrapf(err, "error fetching object %s/%s", writer.s3Bucket, writer.s3Key)
	}
	writer.oldHash = strings.ReplaceAll(*response.ETag, "\"", "")
	return nil
}

//DisableServerSideEncryption overrides and disables the default encryption (AES 256)
func (writer *S3DatabaseWriter) DisableServerSideEncryption() {
	log.Printf("Server side encryption has been disabled for %s/%s", writer.s3Bucket, writer.s3Key)
	writer.disableEncryption = true
}

//Write writes data to temporary file
func (writer *S3DatabaseWriter) Write(p []byte) (n int, err error) {
	return writer.multiWriter.Write(p)
}

//Close closes the temporary file and release the file lock
func (writer *S3DatabaseWriter) Close() error {
	if err := writer.temporaryFile.Close(); err != nil && errors.Cause(err) == os.ErrClosed {
		return errors.Wrap(err, "error closing temporary file")
	}
	if err := os.Remove(writer.temporaryFile.Name()); err != nil && errors.Cause(err) == os.ErrNotExist {
		return errors.Wrap(err, "error removing temporary file")
	}
	if err := writer.lock.Unlock(); err != nil {
		return errors.Wrap(err, "error releasing lock file")
	}
	return nil
}

//ValidHash checks that the temporary file's MD5 matches the expectedHash
func (writer *S3DatabaseWriter) ValidHash(expectedHash string) error {
	actualHash := fmt.Sprintf("%x", writer.md5Writer.Sum(nil))
	if !strings.EqualFold(actualHash, expectedHash) {
		return errors.Errorf("md5 of new database (%s) does not match expected md5 (%s)", actualHash, expectedHash)
	}
	return nil
}

//GetHash returns the hash of the current database file
func (writer *S3DatabaseWriter) GetHash() string {
	return writer.oldHash
}

//SetFileModificationTime explicitly sets the database's file write time to the provided time.  This is stored in the
//	DateOfSourceDatabaseModification tag on the S3 object
func (writer *S3DatabaseWriter) SetFileModificationTime(lastModified time.Time) error {
	tags := make([]*s3.Tag, 1)
	tags[0] = &s3.Tag{
		Key:   aws.String(DateModifiedTag),
		Value: aws.String(lastModified.Format(time.RFC3339)),
	}

	putTagInput := &s3.PutObjectTaggingInput{
		Bucket:  aws.String(writer.s3Bucket),
		Key:     aws.String(writer.s3Key),
		Tagging: &s3.Tagging{TagSet: tags},
	}

	_, err := writer.s3Client.PutObjectTagging(putTagInput)
	if err != nil {
		return errors.Wrap(err, "encountered an error adding modification time tag to S3 object")
	}
	return nil
}

//Commit puts the temporary file into the provided S3 bucket
func (writer *S3DatabaseWriter) Commit() error {
	if err := writer.temporaryFile.Sync(); err != nil {
		return errors.Wrap(err, "error syncing temporary file")
	}
	if err := writer.Close(); err != nil {
		return errors.Wrap(err, "error closing temporary file")
	}
	s3Body, err := os.Open(writer.temporaryFile.Name())
	if err != nil {
		return errors.Wrap(err, "error opening temporary file for S3 Body")
	}
	s3PutObject := &s3.PutObjectInput{
		Bucket:               aws.String(writer.s3Bucket),
		Key:                  aws.String(writer.s3Key),
		ServerSideEncryption: aws.String(Encryption),
		Body:                 s3Body,
	}
	if writer.disableEncryption {
		s3PutObject.ServerSideEncryption = nil
	}
	if _, err := writer.s3Client.PutObject(s3PutObject); err != nil {
		return errors.Wrap(err, "encountered an error writing file to S3")
	}
	return nil
}
