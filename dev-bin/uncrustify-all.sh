#!/bin/sh

uncrustify="uncrustify -c .uncrustify.cfg --replace --no-backup"

# We indent each thing twice because uncrustify is not idempotent - in some
# cases it will flip-flop between two indentation styles.
for dir in bin; do
    c_files=`find $dir -maxdepth 1 -name '*.c' | grep -v 'md5\|base64\|types'`
    echo $c_files
    if [ "$c_files" != "" ]; then
        for file in $c_files; do
            $uncrustify $file
            $uncrustify $file
        done
    fi
    
    h_files=`find $dir -maxdepth 1 -name '*.h' | grep -v 'md5\|base64\|types'`
    echo $h_files;
    if [ "$h_files" != "" ]; then
        for file in $h_files; do
            $uncrustify $file
            $uncrustify $file
        done
    fi
done
