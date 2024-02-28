package parser

import (
	"github.com/benthosdev/benthos/v4/internal/bloblang/mapping"
	"github.com/benthosdev/benthos/v4/internal/bloblang/query"
)

type rootIfStatementPair struct {
	query      query.Function
	statements []mapping.Statement
}

func rootLevelIfExpressionParser(pCtx Context) Func {
	ifBlockStatementsPattern := DelimitedPattern(
		Sequence(
			charSquigOpen,
			DiscardedWhitespaceNewlineComments,
		),
		// Prevent imports, maps and metadata assignments.
		mappingStatement(pCtx, false, nil),
		Sequence(
			Discard(SpacesAndTabs),
			NewlineAllowComment,
			DiscardedWhitespaceNewlineComments,
		),
		Sequence(
			DiscardedWhitespaceNewlineComments,
			charSquigClose,
		),
	)

	return func(input []rune) Result {
		ifParser := Sequence(
			Term("if"),
			SpacesAndTabs,
			queryParser(pCtx),
			DiscardedWhitespaceNewlineComments,
			ifBlockStatementsPattern,
		)

		elseIfParser := Optional(Sequence(
			DiscardedWhitespaceNewlineComments,
			Term("else if"),
			SpacesAndTabs,
			MustBe(queryParser(pCtx)),
			DiscardedWhitespaceNewlineComments,
			ifBlockStatementsPattern,
		))

		elseParser := Optional(Sequence(
			DiscardedWhitespaceNewlineComments,
			Term("else"),
			DiscardedWhitespaceNewlineComments,
			ifBlockStatementsPattern,
		))

		var ifStatements []*rootIfStatementPair

		res := ifParser(input)
		if res.Err != nil {
			return res
		}

		{
			seqSlice := res.Payload.([]any)
			var tmpPair rootIfStatementPair
			tmpPair.query = seqSlice[2].(query.Function)

			stmtSlice := seqSlice[4].([]any)
			tmpPair.statements = make([]mapping.Statement, len(stmtSlice))
			for i, v := range stmtSlice {
				tmpPair.statements[i] = v.(mapping.Statement)
			}
			ifStatements = append(ifStatements, &tmpPair)
		}

		for {
			res = elseIfParser(res.Remaining)
			if res.Err != nil {
				return res
			}
			if res.Payload == nil {
				break
			}
			seqSlice := res.Payload.([]any)

			var tmpPair rootIfStatementPair
			tmpPair.query = seqSlice[3].(query.Function)

			stmtSlice := seqSlice[5].([]any)
			tmpPair.statements = make([]mapping.Statement, len(stmtSlice))
			for i, v := range stmtSlice {
				tmpPair.statements[i] = v.(mapping.Statement)
			}
			ifStatements = append(ifStatements, &tmpPair)
		}

		res = elseParser(res.Remaining)
		if res.Err != nil {
			return res
		}
		if seqSlice, ok := res.Payload.([]any); ok {
			var tmpPair rootIfStatementPair
			stmtSlice := seqSlice[3].([]any)
			tmpPair.statements = make([]mapping.Statement, len(stmtSlice))
			for i, v := range stmtSlice {
				tmpPair.statements[i] = v.(mapping.Statement)
			}
			ifStatements = append(ifStatements, &tmpPair)
		}

		res.Payload = ifStatements
		return res
	}
}
