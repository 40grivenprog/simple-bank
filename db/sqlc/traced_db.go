package db

import (
	"context"
	"database/sql"
	"regexp"

	"github.com/40grivenprog/simple-bank/jaeger"
	"github.com/opentracing/opentracing-go"
)

type TracedDB struct {
	*sql.DB
}

func NewTracedDB(dbDriver, dbSource string) (*TracedDB, error) {
	conn, err := sql.Open(dbDriver, dbSource)
	if err != nil {
		return nil, err
	}

	return &TracedDB{conn}, nil
}

func (tdb *TracedDB) GetConnection() *sql.DB {
	return tdb.DB
}

func (tdb *TracedDB) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	span, ctx := startSpanFromContext(ctx, extractQueryName(query))
	defer span.Finish()

	return tdb.DB.ExecContext(ctx, query, args...)
}

func (tdb *TracedDB) PrepareContext(ctx context.Context, query string) (*sql.Stmt, error) {
	span, ctx := startSpanFromContext(ctx, extractQueryName(query))
	defer span.Finish()

	return tdb.DB.PrepareContext(ctx, query)
}

func (tdb *TracedDB) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	span, ctx := startSpanFromContext(ctx, extractQueryName(query))
	defer span.Finish()

	return tdb.DB.QueryContext(ctx, query, args...)
}

func (tdb *TracedDB) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	span, ctx := startSpanFromContext(ctx, extractQueryName(query))
	defer span.Finish()

	return tdb.DB.QueryRowContext(ctx, query, args...)
}

func startSpanFromContext(ctx context.Context, operationName string) (opentracing.Span, context.Context) {
	span, ok := ctx.Value("span").(opentracing.Span)

	if !ok {
		span = jaeger.Tracer.StartSpan(operationName)
	} else {
		span = jaeger.Tracer.StartSpan(operationName, opentracing.ChildOf(span.Context()))
	}
	return span, opentracing.ContextWithSpan(ctx, span)
}

func extractQueryName(query string) string {
	re := regexp.MustCompile(`-- name: ([\w\s]+)`)
	match := re.FindStringSubmatch(query)

	return match[0]
}
