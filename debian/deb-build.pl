#!/usr/bin/env perl

use strict;
use warnings;

use Cwd qw( cwd );
use Debian::Debhelper::Buildsystem::golang ();

$ENV{PATH} = '/usr/lib/go-1.10/bin:' . $ENV{PATH};

my $version = `dpkg-parsechangelog -SVersion`;
chomp $version;

die 'Version missing!' unless $version;

# See build() in the D::D::B::golang for what we're replicating.
my $b = Debian::Debhelper::Buildsystem::golang->new;
$ENV{GOPATH} = cwd() . '/' . $b->get_builddir;
chdir $b->get_builddir || die $!;

# Hack! xenial fails with an error about missing this file for some reason
# at this stage. Create it to make it happy.
mkdir 'debian' || die $!;
open my $fh, '>', 'debian/compat' || die $!;
print { $fh } "9\n" || die $!;
close $fh || die $!;

# eoan builds fail trying to download Go modules. We vendor them, so use
# -mod=vendor. However, that flag is only supported on Go 1.11+, and versions
# older than eoan use Go 1.10.
my @extra_flags;
my $go_version_raw = `go version`;
chomp $go_version_raw;
my $go_minor_version;
if ($go_version_raw =~ /\Ago version go1\.([0-9]+)(?:\.[0-9]+)?/) {
    $go_minor_version = $1;
}
if (!$go_minor_version || $go_minor_version >= 11) {
    push @extra_flags, '-mod=vendor';
}

system(
    'go',
    'install',
	'-ldflags', "-X main.defaultConfigFile=/etc/GeoIP.conf -X main.defaultDatabaseDirectory=/usr/share/GeoIP -X 'main.version=$version (ubuntu-ppa)'",
    @extra_flags,
    'github.com/maxmind/geoipupdate/...',
) == 0 || die 'error building geoipupdate';

exit 0;
