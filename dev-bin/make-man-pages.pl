#!/usr/bin/env perl
use strict;
use warnings;

use File::Temp qw( tempfile );

sub main {
    my $build_dir = $ARGV[0] // 'build';

    _make_man(
        'geoipupdate',
        1,
        "$build_dir/geoipupdate.md",
        "$build_dir/geoipupdate.1",
    );
    _make_man(
        'GeoIP.conf',
        5,
        "$build_dir/GeoIP.conf.md",
        "$build_dir/GeoIP.conf.5",
    );
    return 1;
}

sub _make_man {
    my ( $name, $section, $input, $output ) = @_;

    my ( $fh, $tmp ) = tempfile();
    binmode $fh or die $!;
    print {$fh} "% $name($section)\n\n" or die $!;
    my $contents = _read($input);
    print {$fh} $contents or die $!;
    close $fh or die $!;

    system(
        'pandoc',
        '-s',
        '-f', 'markdown',
        '-t', 'man',
        $tmp,
        '-o', $output,
    ) == 0 or die 'pandoc failed';

    return;
}

sub _read {
    my ($file) = @_;
    open my $fh, '<', $file or die $!;
    binmode $fh or die $!;
    my $contents = '';
    while ( !eof($fh) ) {
        my $line = <$fh>;
        die 'error reading' unless defined $line;
        $contents .= $line;
    }
    close $fh or die $!;
    return $contents;
}

exit( main() ? 0 : 1 );
