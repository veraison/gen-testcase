// Copyright 2024 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/pflag"
)

const DefaultContentType = "application/rim+cbor"

var outfile *string = pflag.StringP("out", "o", "",
	"Output will be written to this file. If not specified, defaults " +
	"to the same path as the input with the extension changed to .cbor.")

var writeToStdout *bool = pflag.BoolP("stdout", "O", false,
	"Write to standard output instead of a file.")

var signingKey *string = pflag.StringP("signing-key", "s", "",
	"Path to a signing key in JWK format. If this is specified, a COSE " +
	"Sign1Message will be generated with the encoded input as the payload")

var contentType *string = pflag.StringP("contentType", "c", DefaultContentType,
	"When signing with -s/--signing-key, this will be used as the value " +
	"of the content type COSE header.")

var metafile *string = pflag.StringP("meta", "m", "",
	"Path to YAML file that will be encoded and used as the meta header in the " +
	"COSE Sign1Message (when -s/--signing-key is also specified)")

func validateArgs() {
	if pflag.NArg() != 1 {
                log.Fatalf("error: must specify exactly one positional argument")
	}

	if *outfile != "" && *writeToStdout {
                log.Fatalf("error: -o/--out and -O/--stdout cannot be both specified")
	}

	if *signingKey == "" { // not gonna be signing
		if *contentType != DefaultContentType {
			log.Fatalf("error: -c/--content-type should only be used with -s/--signing-key")
		}

		if *metafile != "" {
			log.Fatalf("error: -m/--meta should only be used with -s/--signing-key")
		}
	} else { // gonna be signing
		if *metafile == "" {
			log.Print("warning: generating COSE Sign1Message without -m/--meta")
		}
	}
}

func encodeFileToCBOR(path string) ([]byte, error) {
	data, err := os.ReadFile(path)
        if err != nil {
                return nil, fmt.Errorf("error: %w", err)
        }
	
	return yaml2cbor(data)
}

func main() {
	pflag.Parse()
	validateArgs()

	inFile := pflag.Arg(0)
	out, err := encodeFileToCBOR(inFile)
        if err != nil {
                log.Fatalf("error: %v", err)
        }

	outPath := *outfile
	if outPath == "" {
		outPath = strings.TrimSuffix(inFile, filepath.Ext(inFile)) + ".cbor"
	}

	if *signingKey != "" {
		keyData, err := os.ReadFile(*signingKey)
		if err != nil {
			log.Fatalf("error: %v", err)
		}

		signer, err := coseSignerFromJWK(keyData)
		if err != nil {

			log.Fatalf("error: %v", err)
		}

		var meta []byte
		if *metafile != "" {
			meta, err = encodeFileToCBOR(*metafile)
			if err != nil {
				log.Fatalf("error: %v", err)
			}
		}

		out, err = sign(out, meta, *contentType, signer)
	}

	if outPath == "-" || *writeToStdout {
		fmt.Print(string(out))
	} else {
		err = os.WriteFile(outPath, out, 0666)
		if err != nil {
			log.Fatalf("error: %v", err)
		}
	}
}
