package testpgz

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/georgysavva/scany/sqlscan"
	"github.com/ibrt/golang-errors/errorz"
	"github.com/lensesio/tableprinter"
	"github.com/stretchr/testify/require"

	"github.com/ibrt/golang-inject-pg/pgz"
)

// Profile describes a database profile.
type Profile struct {
	FunctionName    string
	StatementsTotal float64 `db:"statements_total" header:"Statements Total"`
	BranchesTotal   float64 `db:"branches_total" header:"Branches Total"`
	ByLine          []*ProfileByLine
	ByStatement     []*ProfileByStatement
}

// RequireFullCoverage requires full coverage.
func (p *Profile) RequireFullCoverage(t *testing.T) {
	if p.StatementsTotal < 1.0 || p.BranchesTotal < 1.0 {
		p.PrettyPrint()
		require.GreaterOrEqual(t, p.StatementsTotal, 1.0)
		require.GreaterOrEqual(t, p.BranchesTotal, 1.0)
	}
}

// RequireCoverage requires the specified coverage.
func (p *Profile) RequireCoverage(t *testing.T, minStatements, minBranches float64) {
	if p.StatementsTotal < 1.0 || p.BranchesTotal < 1.0 {
		p.PrettyPrint()
		require.GreaterOrEqual(t, p.StatementsTotal, minStatements)
		require.GreaterOrEqual(t, p.BranchesTotal, minBranches)
	}
}

// PrettyPrint pretty prints the Profile.
func (p *Profile) PrettyPrint() {
	w := tableprinter.New(os.Stdout)
	w.BorderTop, w.BorderBottom, w.BorderLeft, w.BorderRight = true, true, true, true
	w.CenterSeparator = "│"
	w.ColumnSeparator = "│"
	w.NumbersAlignment = tableprinter.AlignLeft
	w.RowCharLimit = 1024
	w.RowSeparator = "─"

	fmt.Println()
	fmt.Println("COVERAGE REPORT")
	fmt.Println(p.FunctionName)
	fmt.Println()

	fmt.Println("Coverage Totals")
	w.Print(p)
	fmt.Println()

	fmt.Println("Coverage By Line")
	w.Print(p.ByLine)
	fmt.Println()

	fmt.Println("Coverage By Statement")
	w.Print(p.ByStatement)
	fmt.Println()
}

// ProfileByLine describes a per-line database profile.
type ProfileByLine struct {
	LineNumber     int64  `db:"lineno" header:"Line Number"`
	ExecStatements int64  `db:"exec_stmts" header:"Exec Stmts"`
	Source         string `db:"source" header:"Source"`
}

// ProfileByStatement describes a per-statement database profile.
type ProfileByStatement struct {
	StatementID       int64  `db:"stmtid" header:"Stmt ID"`
	ParentStatementID int64  `db:"parent_stmtid" header:"Parent Stmt ID"`
	ParentNote        string `db:"parent_note" header:"Parent Note"`
	LineNumber        int64  `db:"lineno" header:"Line Number"`
	ExecStatements    int64  `db:"exec_stmts" header:"Exec Stmts"`
	StatementName     string `db:"stmtname" header:"Stmt Name"`
}

// GetProfile returns a database profile for the given function.
func GetProfile(ctx context.Context, functionName string) *Profile {
	profile := &Profile{
		FunctionName:    functionName,
		StatementsTotal: 0,
		BranchesTotal:   0,
		ByLine:          make([]*ProfileByLine, 0),
		ByStatement:     make([]*ProfileByStatement, 0),
	}

	row := pgz.GetCtx(ctx).QueryRow(`SELECT plpgsql_coverage_statements($1)`, functionName)
	errorz.MaybeMustWrap(row.Err(), errorz.Skip())
	errorz.MaybeMustWrap(row.Scan(&profile.StatementsTotal), errorz.Skip())

	row = pgz.GetCtx(ctx).QueryRow(`SELECT plpgsql_coverage_branches($1)`, functionName)
	errorz.MaybeMustWrap(row.Err(), errorz.Skip())
	errorz.MaybeMustWrap(row.Scan(&profile.BranchesTotal), errorz.Skip())

	errorz.MaybeMustWrap(sqlscan.Select(ctx, pgz.Get(ctx), &profile.ByLine, `
		SELECT 
			COALESCE(lineno, 0) AS lineno, 
			COALESCE(exec_stmts, 0) AS exec_stmts, 
			COALESCE(source, '') AS source 
		FROM plpgsql_profiler_function_tb($1)`,
		functionName),
		errorz.Skip())

	errorz.MaybeMustWrap(sqlscan.Select(ctx, pgz.Get(ctx), &profile.ByStatement, `
		SELECT 
			COALESCE(stmtid, 0) AS stmtid, 
			COALESCE(parent_stmtid, 0) AS parent_stmtid, 
			COALESCE(parent_note, '') AS parent_note, 
			COALESCE(lineno, 0) AS lineno,
			COALESCE(exec_stmts, 0) AS exec_stmts, 
			COALESCE(stmtname, '') AS stmtname 
		FROM plpgsql_profiler_function_statements_tb($1)`,
		functionName),
		errorz.Skip())

	return profile
}

// ResetProfiler resets the database profiler data for the given functions.
func ResetProfiler(ctx context.Context, functionNames ...string) {
	for _, functionName := range functionNames {
		_, err := pgz.GetCtx(ctx).Exec(`
			SELECT plpgsql_profiler_reset((
				SELECT oid FROM pg_proc WHERE proname = $1
			))`,
			functionName)
		errorz.MaybeMustWrap(err, errorz.Skip())
		errorz.Assertf(GetProfile(ctx, functionName).StatementsTotal == 0,
			"coverage unexpectedly present", errorz.Skip())
	}
}
