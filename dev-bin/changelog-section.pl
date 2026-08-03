#!/usr/bin/env perl

# Prints the CHANGELOG.md notes for one release.
#
# The GitHub release body is this output: the release workflow writes it to a
# file and passes it to GoReleaser as --release-notes. Without that, GoReleaser
# would generate a commit list instead. dev-bin/release.sh uses the same output
# to preview the notes and to annotate the tag, so both paths agree on what a
# release's notes are.

use strict;
use warnings;

my ($version) = @ARGV;
die "Usage: $0 VERSION\n"
    unless defined $version
    && @ARGV == 1
    && $version =~ /^[0-9]+\.[0-9]+\.[0-9]+(?:-[a-zA-Z0-9.]+)?$/;

my $file = 'CHANGELOG.md';
open my $fh, '<:encoding(UTF-8)', $file or die "$file: $!\n";

# The date is required, not decorative: a section without one has not been
# prepared for release.
my $wanted = qr/^\#\# \Q$version\E \([0-9]{4}-[0-9]{2}-[0-9]{2}\)\s*$/;
my $any    = qr/^\#\# [0-9]+\.[0-9]+\.[0-9]+(?:-[a-zA-Z0-9.]+)?/;

my ( $found, @notes );
while ( my $line = <$fh> ) {
    if ( !$found ) {
        $found = 1 if $line =~ $wanted;
        next;
    }
    last if $line =~ $any;
    push @notes, $line;
}
close $fh or die $!;

die "$file has no dated section for $version!\n" unless $found;

shift @notes while @notes && $notes[0]  =~ /^\s*$/;
pop @notes   while @notes && $notes[-1] =~ /^\s*$/;

die "The $version section of $file is empty!\n" unless @notes;

binmode STDOUT, ':encoding(UTF-8)' or die $!;
print @notes or die $!;
