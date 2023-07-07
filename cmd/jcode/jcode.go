package main

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"flag"
	"fmt"
	"go/token"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/dave/dst"
	"github.com/dave/dst/decorator"

	"github.com/luno/jettison/errors"
	"github.com/luno/jettison/j"
)

const (
	statusError       = 1
	statusHelp        = 2
	statusInvalidFile = 3

	jetPath = "\"github.com/luno/jettison/errors\""
	jPath   = "\"github.com/luno/jettison/j\""
)

var (
	rewrite = flag.Bool("rewrite", false, "rewrite source files")
	cFormat = flag.String("format", "err_hex_16", "error code format")
)

type codeFormat int

const (
	formatBase64 codeFormat = 1
	formatErrHex codeFormat = 2
)

type formatParams struct {
	format codeFormat
	label  string
	gen    func(pkg, variable string) string
	valid  func(pkg, variable, code string) bool
}

var fmtParams = []formatParams{
	{
		formatBase64, "base64_8", fmtBase64, validBase64,
	},
	{
		formatErrHex, "err_hex_16", fmtErrHex, validErrHex,
	},
}

func main() {
	rand.Seed(time.Now().UnixNano())
	flag.Parse()
	if flag.NArg() == 0 {
		fmt.Fprintln(os.Stderr, "No file provided")
		flag.Usage()
		os.Exit(statusHelp)
	}

	fp := getFormatParams()

	fail := false
	for _, file := range flag.Args() {
		res, err := checkFile(file, fp, *rewrite)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%+v\n", err)
			os.Exit(statusError)
		}
		if res.pass {
			continue
		}
		for _, msg := range res.msgs {
			fmt.Println("  ", msg)
		}
		if *rewrite {
			err := os.WriteFile(file, res.out, 0o644)
			if err != nil {
				fmt.Fprintf(os.Stderr, "%+v\n", err)
				os.Exit(statusError)
			}
		} else {
			fail = true
		}
	}
	if fail {
		os.Exit(statusInvalidFile)
	}
}

func getFormatParams() formatParams {
	var fp formatParams
	for _, p := range fmtParams {
		if p.label == *cFormat {
			fp = p
		}
	}
	if fp.label == "" {
		fmt.Fprintln(os.Stderr, "Unknown format")
		flag.Usage()
		os.Exit(statusHelp)
	}
	return fp
}

type checkResult struct {
	pass bool
	msgs []string
	out  []byte
}

func checkFile(file string, fp formatParams, rewrite bool) (checkResult, error) {
	var (
		msgs          []string
		jettisonPkg   string
		pkg           string
		importGenDecl *dst.GenDecl
		foundJ        bool
	)

	fset := token.NewFileSet()
	f, err := decorator.ParseFile(fset, file, nil, 0)
	if err != nil {
		return checkResult{}, err
	}

	for _, decl := range f.Decls {
		gd, ok := decl.(*dst.GenDecl)
		if !ok {
			continue
		}
		if gd.Tok == token.IMPORT {
			// import block
			importGenDecl = gd
			for _, s := range gd.Specs {
				is, ok := s.(*dst.ImportSpec)
				if !ok {
					return checkResult{}, errors.Wrap(err, "Unexpected non import spec",
						j.KV("file", file))
				}
				alias := "errors"
				if is.Name != nil {
					alias = is.Name.Name
				}
				if string(is.Path.Value) == jetPath {
					jettisonPkg = alias
				} else if string(is.Path.Value) == jPath {
					foundJ = true
				}
			}
			continue
		}
		if gd.Tok == token.VAR {
			// var block
			for _, s := range gd.Specs {
				// Check for jettison sentinel errors: var ErrFoo = errors.New("foo").
				vp, ok := s.(*dst.ValueSpec)
				if !ok {
					continue
				}
				if len(vp.Names) != 1 || len(vp.Values) != 1 {
					continue
				}
				ce, ok := vp.Values[0].(*dst.CallExpr)
				if !ok {
					continue
				}
				varName := vp.Names[0].Name
				se, ok := ce.Fun.(*dst.SelectorExpr)
				if !ok {
					continue
				}
				if fmt.Sprint(se.X) != jettisonPkg || se.Sel.Name != "New" {
					continue
				}

				// Check code
				msg, err := checkInstance(jettisonPkg, pkg, varName, ce, fp)
				if err != nil {
					return checkResult{}, errors.Wrap(err, "error checking instance",
						j.MKV{"file": file, "var": varName})
				}
				if msg != "" {
					msgs = append(msgs, fmt.Sprintf("%s: %s: %s (fixed)", file, varName, msg))
				}
			}
		}
	}

	var buf bytes.Buffer
	if len(msgs) > 0 && rewrite {
		if !foundJ {
			// Add jettison/j import
			importGenDecl.Specs = append(importGenDecl.Specs, &dst.ImportSpec{
				Path: &dst.BasicLit{Value: jPath},
			})
		}
		if err := decorator.Fprint(&buf, f); err != nil {
			return checkResult{}, err
		}
	}

	return checkResult{
		pass: len(msgs) == 0,
		msgs: msgs,
		out:  buf.Bytes(),
	}, nil
}

func checkInstance(jettisonPkg string, pkg string, varName string, ce *dst.CallExpr, format formatParams) (string, error) {
	if len(ce.Args) < 2 {
		code := format.gen(pkg, varName)
		ce.Args = append(ce.Args, makeCodeCall(code))
		return "missing jettison code", nil
	}
	if len(ce.Args) > 2 {
		return "", errors.New("only single jettison code option supported")
	}
	ce2, ok := ce.Args[1].(*dst.CallExpr)
	if !ok {
		return "", errors.New("invalid second argument, expect j.C")
	}
	se2, ok := ce2.Fun.(*dst.SelectorExpr)
	if !ok {
		return "", errors.New("invalid second argument, expect j.C")
	}
	selExp := fmt.Sprintf("%s.%s", se2.X, se2.Sel)
	if selExp != jettisonPkg+".WithCode" && selExp != "j.C" {
		return "", errors.New("invalid second argument, expect j.C")
	}

	bl, ok := ce2.Args[0].(*dst.BasicLit)
	if !ok {
		return "", errors.New("invalid code argument, expect string")
	}

	if !format.valid(pkg, varName, strings.Trim(bl.Value, "\"")) {
		code := format.gen(pkg, varName)
		ce.Args[1] = makeCodeCall(code)
		return "incorrect jettison code", nil
	}

	return "", nil
}

func makeCodeCall(code string) *dst.CallExpr {
	codeExp := &dst.SelectorExpr{
		X:   dst.NewIdent("j"),
		Sel: dst.NewIdent("C"),
	}
	return &dst.CallExpr{
		Fun:  codeExp,
		Args: []dst.Expr{dst.NewIdent(fmt.Sprintf("\"%s\"", code))},
	}
}

// fmtBase64 returns a base64 string of length 8; which is equivalent to a random uint48.
func fmtBase64(_, _ string) string {
	b := make([]byte, 6)
	_, err := rand.Read(b)
	if err != nil {
		return "sorry error"
	}
	return base64.RawURLEncoding.EncodeToString(b)
}

func validBase64(_, _, code string) bool {
	if len(code) != 8 {
		return false
	}
	_, err := base64.RawURLEncoding.DecodeString(code)
	return err == nil
}

// fmtErrHex returns string that matches the following regex "ERR_[0-9a-f]{16}",
// ie. "ERR_" followed by random 16 character hex string.
func fmtErrHex(_, _ string) string {
	b := make([]byte, 8)
	_, err := rand.Read(b)
	if err != nil {
		return "sorry error"
	}
	return fmt.Sprintf("ERR_%x", b)
}

func validErrHex(_, _, code string) bool {
	s := strings.Split(code, "_")
	if len(s) != 2 {
		return false
	}
	if s[0] != "ERR" {
		return false
	}
	_, err := hex.DecodeString(s[1])
	return err == nil
}
