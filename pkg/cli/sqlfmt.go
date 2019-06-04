// Copyright 2018 The Cockroach Authors.
//
// Use of this software is governed by the Business Source License included
// in the file licenses/BSL.txt and at www.mariadb.com/bsl11.
//
// Change Date: 2022-10-01
//
// On the date above, in accordance with the Business Source License, use
// of this software will be governed by the Apache License, Version 2.0,
// included in the file licenses/APL.txt and at
// https://www.apache.org/licenses/LICENSE-2.0

package cli

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/cockroachdb/cockroach/pkg/sql/parser"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// TODO(mjibson): This subcommand has more flags than I would prefer. My
// goal is to have it have just -e and nothing else. gofmt initially started
// with tabs/spaces and width specifiers but later regretted that decision
// and removed them. I would like to get to the point where we achieve SQL
// formatting nirvana and we make this an opinionated formatter with few or
// zero options, hopefully before 2.1.

var sqlfmtCmd = &cobra.Command{
	Use:   "sqlfmt",
	Short: "format SQL statements",
	Long:  "Formats SQL statements from stdin to line length n.",
	RunE:  runSQLFmt,
}

func runSQLFmt(cmd *cobra.Command, args []string) error {
	if sqlfmtCtx.len < 1 {
		return errors.Errorf("line length must be > 0: %d", sqlfmtCtx.len)
	}
	if sqlfmtCtx.tabWidth < 1 {
		return errors.Errorf("tab width must be > 0: %d", sqlfmtCtx.tabWidth)
	}

	var sl parser.Statements
	if len(sqlfmtCtx.execStmts) != 0 {
		for _, exec := range sqlfmtCtx.execStmts {
			stmts, err := parser.Parse(exec)
			if err != nil {
				return err
			}
			sl = append(sl, stmts...)
		}
	} else {
		in, err := ioutil.ReadAll(os.Stdin)
		if err != nil {
			return err
		}
		sl, err = parser.Parse(string(in))
		if err != nil {
			return err
		}
	}

	cfg := tree.DefaultPrettyCfg()
	cfg.UseTabs = !sqlfmtCtx.useSpaces
	cfg.LineWidth = sqlfmtCtx.len
	cfg.TabWidth = sqlfmtCtx.tabWidth
	cfg.Simplify = !sqlfmtCtx.noSimplify
	cfg.Align = tree.PrettyNoAlign
	if sqlfmtCtx.align {
		cfg.Align = tree.PrettyAlignAndDeindent
	}

	for i := range sl {
		fmt.Print(cfg.Pretty(sl[i].AST))
		if len(sl) > 1 {
			fmt.Print(";")
		}
		fmt.Println()
	}
	return nil
}
