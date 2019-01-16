#!/usr/bin/env perl

use strict;
use warnings;

use Cwd qw( cwd );
use Debian::Debhelper::Buildsystem::golang ();

# See build() in the D::D::B::golang for what we're replicating.
my $b = Debian::Debhelper::Buildsystem::golang->new;
$ENV{GOPATH} = cwd() . '/' . $b->get_builddir;
chdir $b->get_builddir || die $!;
system(
    'go',
    'install',
	'-ldflags', '-X main.defaultConfigFile=/etc/GeoIP.conf -X main.defaultDatabaseDirectory=/usr/share/GeoIP',
    'github.com/maxmind/geoipupdate/...',
) == 0 || die 'error building geoipupdate';

exit 0;
