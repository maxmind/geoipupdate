#!/usr/bin/perl

=pod

/*
 *
 * Copyright (C) 2008 MaxMind LLC
 *
 * This library is free software; you can redistribute it and/or
 * modify it under the terms of the GNU Lesser General Public
 * License as published by the Free Software Foundation; either
 * version 2.1 of the License, or (at your option) any later version.
 * 
 * This library is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU
 * Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public
 * License along with this library; if not, write to the Free Software
 * Foundation, Inc., 51 Franklin Street, Fifth Floor, Boston, MA  02110-1301  USA
 */

=cut

=pod

pure perl version of geoipupdate. can handle anything, that 

  GeoIP_update_database
  GeoIP_update_database_general
  
handle. It is a drop in replacement for geoipupdate, as opposide to geoipupdate is the
pp version able to handle proxy requests even with authentication and can be used with
https

=cut

use strict;
use warnings;

our $VERSION = '0.08';

use 5.008;
use Data::Dumper;
use Digest::MD5;
use File::Spec;
use File::Basename;
use Getopt::Std;
use HTTP::Request::Common;
use LWP::UserAgent;
use PerlIO::gzip;
use URI;

my $ua = LWP::UserAgent->new( agent => "pp_geoipupdate/$VERSION" );
$ua->env_proxy;

## --- for auth proxies use
## $ua->proxy(['http', 'ftp'] => 'http://username:password@proxy.myorg.com');

my $license_file = 'GeoIP.conf';
my $update_host  = 'updates.maxmind.com';
my $proto        = 'http';
my %opts;

if ( !getopts( 'hvf:d:', \%opts ) or $opts{h} ) {
    print STDERR
        "Usage: geoipupdate [-hv] [-f license_file] [-d custom directory]\n";
    exit @ARGV ? 1 : 0;
}

my $rootdir = File::Spec->rootdir;
$opts{d} ||= File::Spec->catfile( $rootdir, qw/ usr local share GeoIP / );
$opts{f}
    ||= File::Spec->catfile( $rootdir, qw/ usr local etc /, $license_file );

die "dir $opts{d} does not exist or is not readable or is not a directory\n"
    unless -d $opts{d};
die "license_file $opts{f} does not exist, is not readable or is not a file\n"
    unless -f $opts{f};

#
# --- parse license file
#
open my $fh, '<', $opts{f}
    or die "Error opening GeoIP Configuration file $opts{f}\n";
print "Opened License file $opts{f}\n" if $opts{v};

my ( $user_id, $license_key, @product_ids );
{
    local $_;

    while (<$fh>) {
        next if /^\s*#/;    # skip comments
        /^\s*UserId\s+(\d+)/        and $user_id     = $1, next;
        /^\s*LicenseKey\s+(\S{12})/ and $license_key = $1, next;
        /^\s*ProductIds\s+(\d+(?:[a-zA-Z]{2,3})?(?:\s+\d+(?:[a-zA-Z]{2,3})?)*)/
            and @product_ids = split( /\s+/, $1 ), next;

    }
}

if ( $opts{v} ) {
    print "User id $user_id\n" if $user_id;
    print "Read in license key $license_key\n";
    print "Product ids @product_ids\n";
}

my $err_cnt = 0;

my $print_on_error = sub {
    my $err = shift;
    return unless $err;
    if ( $err !~ /^No new updates available/i ) {
        print STDERR $err, $/;
        $err_cnt++;
    }
    else {
        print $err;
    }
};

if ($user_id) {
    for my $product_id (@product_ids) {

        # update the databases using the user id string,
        # the license key string and the product id for each database
        eval {
            GeoIP_update_database_general(
                $user_id,    $license_key,
                $product_id, $opts{v}
            );
        };
        $print_on_error->($@);
    }
}
else {

    # Old format with just license key for MaxMind GeoIP Country database updates
    # here for backwards compatibility
    eval { GeoIP_update_database( $license_key, $opts{v} ); };
    $print_on_error->($@);
}

exit( $err_cnt > 0 ? 1 : 0 );

sub GeoIP_update_database_general {
    my ( $user_id, $license_key, $product_id, $verbose, $client_ipaddr ) = @_;
    my $u = URI->new("$proto://$update_host/app/update_getfilename");
    $u->query_form( product_id => $product_id );

    print 'Send request ' . $u->as_string, "\n" if ($verbose);
    my $res = $ua->request( GET $u->as_string, Host => $update_host );
    die $res->status_line unless ( $res->is_success );

    # make sure to use only the filename for security reason
    my $geoip_filename
        = File::Spec->catfile( $opts{d}, basename( $res->content ) );

    # /* get MD5 of current GeoIP database file */
    my $old_md5 = _get_hexdigest($geoip_filename);

    print "MD5 sum of database $geoip_filename is $old_md5\n" if $verbose;

    unless ($client_ipaddr) {
        print 'Send request ' . $u->as_string, "\n" if ($verbose);

        # /* get client ip address from MaxMind web page */
        $res = $ua->request(
            GET "$proto://$update_host/app/update_getipaddr",
            Host => $update_host
        );
        die $res->status_line unless ( $res->is_success );
        $client_ipaddr = $res->content;
    }

    print "client ip address: $client_ipaddr\n" if $verbose;
    my $hex_digest2
        = Digest::MD5->new->add( $license_key, $client_ipaddr )->hexdigest;
    print "md5sum of ip address and license key is $hex_digest2\n"
        if $verbose;

    my $mk_db_req_cref = sub {

        $u->path('/app/update_secure');
        $u->query_form(
            db_md5        => shift,
            challenge_md5 => $hex_digest2,
            user_id       => $user_id,
            edition_id    => $product_id
        );
        print 'Send request ' . $u->as_string, "\n" if ($verbose);
        return $ua->request( GET $u->as_string, Host => $update_host );
    };
    $res = $mk_db_req_cref->($old_md5);
    die $res->status_line unless ( $res->is_success );

    # print Dumper($res);
    print "Downloading gzipped GeoIP Database...\n" if $verbose;

    _gunzip_and_replace(
        $res->content,
        $geoip_filename,
        sub {

            # as sanity check request a update for the new downloaded file
            # md5 of the new unpacked file
            my $new_md5 = _get_hexdigest(shift);
            return $mk_db_req_cref->($new_md5);
        }
    );
    print "Done\n" if $verbose;
}

sub GeoIP_update_database {
    my ( $license_key, $verbose ) = @_;
    my $geoip_filename = File::Spec->catfile( $opts{d}, 'GeoIP.dat' );

    # /* get MD5 of current GeoIP database file */
    my $hexdigest = _get_hexdigest($geoip_filename);

    print "MD5 sum of database $geoip_filename is $hexdigest\n" if $verbose;

    my $u = URI->new("$proto://$update_host/app/update");
    $u->query_form( license_key => $license_key, md5 => $hexdigest );

    print 'Send request ' . $u->as_string, "\n" if ($verbose);
    my $res = $ua->request( GET $u->as_string, Host => $update_host );
    die $res->status_line unless ( $res->is_success );
    print "Downloading gzipped GeoIP Database...\n" if $verbose;
    _gunzip_and_replace( $res->content, $geoip_filename );
    print "Done\n" if $verbose;

}

# --- hexdigest of the file or 00000000000000000000000000000000
sub _get_hexdigest {
    my $md5 = '0' x 32;
    if ( open my $fh, '<:raw', shift ) {
        $md5 = Digest::MD5->new->addfile($fh)->hexdigest;
    }
    return $md5;
}

sub _gunzip_and_replace {
    my ( $content, $geoip_filename, $sanity_check_c ) = @_;
    my $max_retry = 1;

    my $tmp_fname = $geoip_filename . '.test';

    {

        # --- error if our content does not start with the gzip header
        die $content || 'Not a gzip file'
            if substr( $content, 0, 2 ) ne "\x1f\x8b";

        # --- uncompress the gzip data
        {
            local $_;
            open my $gin,  '<:gzip', \$content  or die $!;
            open my $gout, '>:raw',  $tmp_fname or die $!;
            print {$gout} $_ while (<$gin>);
        }

        # --- sanity check
        if ( defined $sanity_check_c ) {
            die "Download failed" if $max_retry-- <= 0;
            my $res = $sanity_check_c->($tmp_fname);
            die $res->status_line unless ( $res->is_success );
            $content = $res->content;

            redo if ( $content !~ /^No new updates available/ );
        }
    }

    # --- install GeoIP.dat.test -> GeoIP.dat
    rename( $tmp_fname, $geoip_filename ) or die $!;
}

