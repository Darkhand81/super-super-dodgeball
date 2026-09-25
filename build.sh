#!/usr/bin/env sh
export PATH=$PATH:../cc65-snapshot-win32/bin
set -e

cd "$(dirname "$0")"

# Headered Super Dodge Ball (U) rom used to generate the CHR ROM tiles.
# Override with: NES_ROM=/path/to/rom.nes ./build.sh
NES_ROM="${NES_ROM:-utilities/Super Dodge Ball (U).nes}"

# Generate the chrom-tiles-X.asm files if any are missing
for i in 0 1 2 3 4 5 6 7; do
    if [ ! -f "./src/chrom-tiles-$i.asm" ]; then
        echo "Generating CHR ROM tiles from $NES_ROM"
        go run ./utilities/parseNesFileToBanks.go -in="$NES_ROM" -out=./src
        break
    fi
done

mkdir -p out
ca65 ./src/main.asm -o ./out/main.o -g
ld65 -C ./src/hirom.cfg -o ./out/super-super-dodge-ball.sfc ./out/main.o
