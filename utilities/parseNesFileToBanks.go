package main

import (
	"bufio"
	"crypto/md5"
	"encoding/hex"
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

// md5 of the headered Super Dodge Ball (U) rom this port is built against
const expectedMD5 = "9c819e679f5fab4ef836761d31e98adc"

const (
	headerSize = 0x10
	bankSize   = 0x4000
	// banks 0-7 are PRG ROM, banks 8-15 are CHR ROM
	firstChrBank = 8
	chrBankCount = 8
)

// This utility will take the Super Dodge Ball NES rom as input and generate the
// chrom-tiles-X.asm files the SNES port needs.  It assumes a lot about the game:
// Super Dodge Ball is a MMC1 game, where banks 8+ are CHR ROM banks and within those banks
// banks at memory 1A000 - 1FFFF are data and 8000 - 19FFF are tiles.  It will automatically
// convert the 2bpp tiles into 4bpp SNES format, but leave the data parts as is.
//
// The NES PRG banks are NOT written out: src/bankX.asm are heavily modified for the
// port and must not be replaced with the raw NES code.
//
// This script also assumes that the file will have a 16 byte header that we skip.
//
// It very naively will just print out the code as:
//
// .byte $HL, $HL, ......
//
// with 16 bytes per line
func main() {
	inputFile := flag.String("in", "Super Dodge Ball (U).nes", "headered Super Dodge Ball (U) NES rom")
	outDir := flag.String("out", "../src", "directory to write the chrom-tiles-X.asm files to")
	skipMD5 := flag.Bool("skip-md5", false, "don't verify the input rom's md5")
	flag.Parse()

	if err := run(*inputFile, *outDir, *skipMD5); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func run(inputFile, outDir string, skipMD5 bool) error {
	inputBytes, err := os.ReadFile(inputFile)
	if err != nil {
		return err
	}

	if !skipMD5 {
		sum := md5.Sum(inputBytes)
		if got := hex.EncodeToString(sum[:]); got != expectedMD5 {
			return fmt.Errorf("%s has md5 %s, expected %s (headered Super Dodge Ball (U)); use -skip-md5 to ignore",
				inputFile, got, expectedMD5)
		}
	}

	expectedSize := headerSize + (firstChrBank+chrBankCount)*bankSize
	if len(inputBytes) < expectedSize {
		return fmt.Errorf("%s is %d bytes, expected at least %d", inputFile, len(inputBytes), expectedSize)
	}

	// remove the header
	headerLess := inputBytes[headerSize:]

	var banks [][]byte
	for i := 0; i+bankSize <= len(headerLess); i += bankSize {
		banks = append(banks, headerLess[i:i+bankSize])
	}

	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return err
	}

	// CHR banks
	tileset := 0
	for i := firstChrBank; i < firstChrBank+chrBankCount; i++ {
		path := filepath.Join(outDir, fmt.Sprintf("chrom-tiles-%d.asm", i-firstChrBank))
		if err := writeChrBank(path, banks[i], i, &tileset); err != nil {
			return err
		}
		fmt.Println("wrote", path)
	}

	return nil
}

func writeChrBank(path string, bank []byte, i int, tileset *int) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	w := bufio.NewWriter(file)
	fmt.Fprintf(w, ".segment \"PRGA%X\"\n", i)
	for byteIndex := 0; byteIndex < len(bank); byteIndex += 0x10 {
		if byteIndex%0x1000 == 0 {
			fmt.Fprintf(w, "chrom_bank_%d_tileset_%d:\n", i-firstChrBank, *tileset)
			*tileset++
		}

		if i < 14 || (byteIndex < 0x2000 && i == 14) {
			// converts these to SNES expected format
			fmt.Fprintf(w,
				".byte $%02X, $%02X, $%02X, $%02X, $%02X, $%02X, $%02X, $%02X, $%02X,"+
					" $%02X, $%02X, $%02X, $%02X, $%02X, $%02X, $%02X\n",
				bank[byteIndex],
				bank[byteIndex+8],
				bank[byteIndex+1],
				bank[byteIndex+1+8],
				bank[byteIndex+2],
				bank[byteIndex+2+8],
				bank[byteIndex+3],
				bank[byteIndex+3+8],
				bank[byteIndex+4],
				bank[byteIndex+4+8],
				bank[byteIndex+5],
				bank[byteIndex+5+8],
				bank[byteIndex+6],
				bank[byteIndex+6+8],
				bank[byteIndex+7],
				bank[byteIndex+7+8],
			)
			w.WriteString(".byte $00, $00, $00, $00, $00, $00, $00, $00, $00, $00, $00, $00, $00, $00, $00, $00\n")
		} else {
			// these are data banks that need to be formatted differently
			fmt.Fprintf(w,
				".byte $%02X, $00, $%02X, $00, $%02X, $00, $%02X, $00, $%02X, $00, $%02X, $00, $%02X, $00, $%02X, $00\n"+
					".byte $%02X, $00, $%02X, $00, $%02X, $00, $%02X, $00, $%02X, $00, $%02X, $00, $%02X, $00, $%02X, $00\n",
				bank[byteIndex],
				bank[byteIndex+1],
				bank[byteIndex+2],
				bank[byteIndex+3],
				bank[byteIndex+4],
				bank[byteIndex+5],
				bank[byteIndex+6],
				bank[byteIndex+7],
				bank[byteIndex+8],
				bank[byteIndex+9],
				bank[byteIndex+10],
				bank[byteIndex+11],
				bank[byteIndex+12],
				bank[byteIndex+13],
				bank[byteIndex+14],
				bank[byteIndex+15],
			)
		}
	}

	if err := w.Flush(); err != nil {
		return err
	}
	return file.Close()
}
