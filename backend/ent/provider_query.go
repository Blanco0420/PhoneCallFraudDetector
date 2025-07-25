// Code generated by ent, DO NOT EDIT.

package ent

import (
	"context"
	"database/sql/driver"
	"fmt"
	"math"

	"entgo.io/ent"
	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/schema/field"
	"github.com/Blanco0420/Phone-Number-Check/backend/ent/business"
	"github.com/Blanco0420/Phone-Number-Check/backend/ent/comment"
	"github.com/Blanco0420/Phone-Number-Check/backend/ent/number"
	"github.com/Blanco0420/Phone-Number-Check/backend/ent/predicate"
	"github.com/Blanco0420/Phone-Number-Check/backend/ent/provider"
)

// ProviderQuery is the builder for querying Provider entities.
type ProviderQuery struct {
	config
	ctx          *QueryContext
	order        []provider.OrderOption
	inters       []Interceptor
	predicates   []predicate.Provider
	withNumber   *NumberQuery
	withComment  *CommentQuery
	withBusiness *BusinessQuery
	withFKs      bool
	// intermediate query (i.e. traversal path).
	sql  *sql.Selector
	path func(context.Context) (*sql.Selector, error)
}

// Where adds a new predicate for the ProviderQuery builder.
func (pq *ProviderQuery) Where(ps ...predicate.Provider) *ProviderQuery {
	pq.predicates = append(pq.predicates, ps...)
	return pq
}

// Limit the number of records to be returned by this query.
func (pq *ProviderQuery) Limit(limit int) *ProviderQuery {
	pq.ctx.Limit = &limit
	return pq
}

// Offset to start from.
func (pq *ProviderQuery) Offset(offset int) *ProviderQuery {
	pq.ctx.Offset = &offset
	return pq
}

// Unique configures the query builder to filter duplicate records on query.
// By default, unique is set to true, and can be disabled using this method.
func (pq *ProviderQuery) Unique(unique bool) *ProviderQuery {
	pq.ctx.Unique = &unique
	return pq
}

// Order specifies how the records should be ordered.
func (pq *ProviderQuery) Order(o ...provider.OrderOption) *ProviderQuery {
	pq.order = append(pq.order, o...)
	return pq
}

// QueryNumber chains the current query on the "number" edge.
func (pq *ProviderQuery) QueryNumber() *NumberQuery {
	query := (&NumberClient{config: pq.config}).Query()
	query.path = func(ctx context.Context) (fromU *sql.Selector, err error) {
		if err := pq.prepareQuery(ctx); err != nil {
			return nil, err
		}
		selector := pq.sqlQuery(ctx)
		if err := selector.Err(); err != nil {
			return nil, err
		}
		step := sqlgraph.NewStep(
			sqlgraph.From(provider.Table, provider.FieldID, selector),
			sqlgraph.To(number.Table, number.FieldID),
			sqlgraph.Edge(sqlgraph.M2O, true, provider.NumberTable, provider.NumberColumn),
		)
		fromU = sqlgraph.SetNeighbors(pq.driver.Dialect(), step)
		return fromU, nil
	}
	return query
}

// QueryComment chains the current query on the "comment" edge.
func (pq *ProviderQuery) QueryComment() *CommentQuery {
	query := (&CommentClient{config: pq.config}).Query()
	query.path = func(ctx context.Context) (fromU *sql.Selector, err error) {
		if err := pq.prepareQuery(ctx); err != nil {
			return nil, err
		}
		selector := pq.sqlQuery(ctx)
		if err := selector.Err(); err != nil {
			return nil, err
		}
		step := sqlgraph.NewStep(
			sqlgraph.From(provider.Table, provider.FieldID, selector),
			sqlgraph.To(comment.Table, comment.FieldID),
			sqlgraph.Edge(sqlgraph.O2M, false, provider.CommentTable, provider.CommentColumn),
		)
		fromU = sqlgraph.SetNeighbors(pq.driver.Dialect(), step)
		return fromU, nil
	}
	return query
}

// QueryBusiness chains the current query on the "business" edge.
func (pq *ProviderQuery) QueryBusiness() *BusinessQuery {
	query := (&BusinessClient{config: pq.config}).Query()
	query.path = func(ctx context.Context) (fromU *sql.Selector, err error) {
		if err := pq.prepareQuery(ctx); err != nil {
			return nil, err
		}
		selector := pq.sqlQuery(ctx)
		if err := selector.Err(); err != nil {
			return nil, err
		}
		step := sqlgraph.NewStep(
			sqlgraph.From(provider.Table, provider.FieldID, selector),
			sqlgraph.To(business.Table, business.FieldID),
			sqlgraph.Edge(sqlgraph.O2O, false, provider.BusinessTable, provider.BusinessColumn),
		)
		fromU = sqlgraph.SetNeighbors(pq.driver.Dialect(), step)
		return fromU, nil
	}
	return query
}

// First returns the first Provider entity from the query.
// Returns a *NotFoundError when no Provider was found.
func (pq *ProviderQuery) First(ctx context.Context) (*Provider, error) {
	nodes, err := pq.Limit(1).All(setContextOp(ctx, pq.ctx, ent.OpQueryFirst))
	if err != nil {
		return nil, err
	}
	if len(nodes) == 0 {
		return nil, &NotFoundError{provider.Label}
	}
	return nodes[0], nil
}

// FirstX is like First, but panics if an error occurs.
func (pq *ProviderQuery) FirstX(ctx context.Context) *Provider {
	node, err := pq.First(ctx)
	if err != nil && !IsNotFound(err) {
		panic(err)
	}
	return node
}

// FirstID returns the first Provider ID from the query.
// Returns a *NotFoundError when no Provider ID was found.
func (pq *ProviderQuery) FirstID(ctx context.Context) (id int, err error) {
	var ids []int
	if ids, err = pq.Limit(1).IDs(setContextOp(ctx, pq.ctx, ent.OpQueryFirstID)); err != nil {
		return
	}
	if len(ids) == 0 {
		err = &NotFoundError{provider.Label}
		return
	}
	return ids[0], nil
}

// FirstIDX is like FirstID, but panics if an error occurs.
func (pq *ProviderQuery) FirstIDX(ctx context.Context) int {
	id, err := pq.FirstID(ctx)
	if err != nil && !IsNotFound(err) {
		panic(err)
	}
	return id
}

// Only returns a single Provider entity found by the query, ensuring it only returns one.
// Returns a *NotSingularError when more than one Provider entity is found.
// Returns a *NotFoundError when no Provider entities are found.
func (pq *ProviderQuery) Only(ctx context.Context) (*Provider, error) {
	nodes, err := pq.Limit(2).All(setContextOp(ctx, pq.ctx, ent.OpQueryOnly))
	if err != nil {
		return nil, err
	}
	switch len(nodes) {
	case 1:
		return nodes[0], nil
	case 0:
		return nil, &NotFoundError{provider.Label}
	default:
		return nil, &NotSingularError{provider.Label}
	}
}

// OnlyX is like Only, but panics if an error occurs.
func (pq *ProviderQuery) OnlyX(ctx context.Context) *Provider {
	node, err := pq.Only(ctx)
	if err != nil {
		panic(err)
	}
	return node
}

// OnlyID is like Only, but returns the only Provider ID in the query.
// Returns a *NotSingularError when more than one Provider ID is found.
// Returns a *NotFoundError when no entities are found.
func (pq *ProviderQuery) OnlyID(ctx context.Context) (id int, err error) {
	var ids []int
	if ids, err = pq.Limit(2).IDs(setContextOp(ctx, pq.ctx, ent.OpQueryOnlyID)); err != nil {
		return
	}
	switch len(ids) {
	case 1:
		id = ids[0]
	case 0:
		err = &NotFoundError{provider.Label}
	default:
		err = &NotSingularError{provider.Label}
	}
	return
}

// OnlyIDX is like OnlyID, but panics if an error occurs.
func (pq *ProviderQuery) OnlyIDX(ctx context.Context) int {
	id, err := pq.OnlyID(ctx)
	if err != nil {
		panic(err)
	}
	return id
}

// All executes the query and returns a list of Providers.
func (pq *ProviderQuery) All(ctx context.Context) ([]*Provider, error) {
	ctx = setContextOp(ctx, pq.ctx, ent.OpQueryAll)
	if err := pq.prepareQuery(ctx); err != nil {
		return nil, err
	}
	qr := querierAll[[]*Provider, *ProviderQuery]()
	return withInterceptors[[]*Provider](ctx, pq, qr, pq.inters)
}

// AllX is like All, but panics if an error occurs.
func (pq *ProviderQuery) AllX(ctx context.Context) []*Provider {
	nodes, err := pq.All(ctx)
	if err != nil {
		panic(err)
	}
	return nodes
}

// IDs executes the query and returns a list of Provider IDs.
func (pq *ProviderQuery) IDs(ctx context.Context) (ids []int, err error) {
	if pq.ctx.Unique == nil && pq.path != nil {
		pq.Unique(true)
	}
	ctx = setContextOp(ctx, pq.ctx, ent.OpQueryIDs)
	if err = pq.Select(provider.FieldID).Scan(ctx, &ids); err != nil {
		return nil, err
	}
	return ids, nil
}

// IDsX is like IDs, but panics if an error occurs.
func (pq *ProviderQuery) IDsX(ctx context.Context) []int {
	ids, err := pq.IDs(ctx)
	if err != nil {
		panic(err)
	}
	return ids
}

// Count returns the count of the given query.
func (pq *ProviderQuery) Count(ctx context.Context) (int, error) {
	ctx = setContextOp(ctx, pq.ctx, ent.OpQueryCount)
	if err := pq.prepareQuery(ctx); err != nil {
		return 0, err
	}
	return withInterceptors[int](ctx, pq, querierCount[*ProviderQuery](), pq.inters)
}

// CountX is like Count, but panics if an error occurs.
func (pq *ProviderQuery) CountX(ctx context.Context) int {
	count, err := pq.Count(ctx)
	if err != nil {
		panic(err)
	}
	return count
}

// Exist returns true if the query has elements in the graph.
func (pq *ProviderQuery) Exist(ctx context.Context) (bool, error) {
	ctx = setContextOp(ctx, pq.ctx, ent.OpQueryExist)
	switch _, err := pq.FirstID(ctx); {
	case IsNotFound(err):
		return false, nil
	case err != nil:
		return false, fmt.Errorf("ent: check existence: %w", err)
	default:
		return true, nil
	}
}

// ExistX is like Exist, but panics if an error occurs.
func (pq *ProviderQuery) ExistX(ctx context.Context) bool {
	exist, err := pq.Exist(ctx)
	if err != nil {
		panic(err)
	}
	return exist
}

// Clone returns a duplicate of the ProviderQuery builder, including all associated steps. It can be
// used to prepare common query builders and use them differently after the clone is made.
func (pq *ProviderQuery) Clone() *ProviderQuery {
	if pq == nil {
		return nil
	}
	return &ProviderQuery{
		config:       pq.config,
		ctx:          pq.ctx.Clone(),
		order:        append([]provider.OrderOption{}, pq.order...),
		inters:       append([]Interceptor{}, pq.inters...),
		predicates:   append([]predicate.Provider{}, pq.predicates...),
		withNumber:   pq.withNumber.Clone(),
		withComment:  pq.withComment.Clone(),
		withBusiness: pq.withBusiness.Clone(),
		// clone intermediate query.
		sql:  pq.sql.Clone(),
		path: pq.path,
	}
}

// WithNumber tells the query-builder to eager-load the nodes that are connected to
// the "number" edge. The optional arguments are used to configure the query builder of the edge.
func (pq *ProviderQuery) WithNumber(opts ...func(*NumberQuery)) *ProviderQuery {
	query := (&NumberClient{config: pq.config}).Query()
	for _, opt := range opts {
		opt(query)
	}
	pq.withNumber = query
	return pq
}

// WithComment tells the query-builder to eager-load the nodes that are connected to
// the "comment" edge. The optional arguments are used to configure the query builder of the edge.
func (pq *ProviderQuery) WithComment(opts ...func(*CommentQuery)) *ProviderQuery {
	query := (&CommentClient{config: pq.config}).Query()
	for _, opt := range opts {
		opt(query)
	}
	pq.withComment = query
	return pq
}

// WithBusiness tells the query-builder to eager-load the nodes that are connected to
// the "business" edge. The optional arguments are used to configure the query builder of the edge.
func (pq *ProviderQuery) WithBusiness(opts ...func(*BusinessQuery)) *ProviderQuery {
	query := (&BusinessClient{config: pq.config}).Query()
	for _, opt := range opts {
		opt(query)
	}
	pq.withBusiness = query
	return pq
}

// GroupBy is used to group vertices by one or more fields/columns.
// It is often used with aggregate functions, like: count, max, mean, min, sum.
//
// Example:
//
//	var v []struct {
//		Name string `json:"Name,omitempty"`
//		Count int `json:"count,omitempty"`
//	}
//
//	client.Provider.Query().
//		GroupBy(provider.FieldName).
//		Aggregate(ent.Count()).
//		Scan(ctx, &v)
func (pq *ProviderQuery) GroupBy(field string, fields ...string) *ProviderGroupBy {
	pq.ctx.Fields = append([]string{field}, fields...)
	grbuild := &ProviderGroupBy{build: pq}
	grbuild.flds = &pq.ctx.Fields
	grbuild.label = provider.Label
	grbuild.scan = grbuild.Scan
	return grbuild
}

// Select allows the selection one or more fields/columns for the given query,
// instead of selecting all fields in the entity.
//
// Example:
//
//	var v []struct {
//		Name string `json:"Name,omitempty"`
//	}
//
//	client.Provider.Query().
//		Select(provider.FieldName).
//		Scan(ctx, &v)
func (pq *ProviderQuery) Select(fields ...string) *ProviderSelect {
	pq.ctx.Fields = append(pq.ctx.Fields, fields...)
	sbuild := &ProviderSelect{ProviderQuery: pq}
	sbuild.label = provider.Label
	sbuild.flds, sbuild.scan = &pq.ctx.Fields, sbuild.Scan
	return sbuild
}

// Aggregate returns a ProviderSelect configured with the given aggregations.
func (pq *ProviderQuery) Aggregate(fns ...AggregateFunc) *ProviderSelect {
	return pq.Select().Aggregate(fns...)
}

func (pq *ProviderQuery) prepareQuery(ctx context.Context) error {
	for _, inter := range pq.inters {
		if inter == nil {
			return fmt.Errorf("ent: uninitialized interceptor (forgotten import ent/runtime?)")
		}
		if trv, ok := inter.(Traverser); ok {
			if err := trv.Traverse(ctx, pq); err != nil {
				return err
			}
		}
	}
	for _, f := range pq.ctx.Fields {
		if !provider.ValidColumn(f) {
			return &ValidationError{Name: f, err: fmt.Errorf("ent: invalid field %q for query", f)}
		}
	}
	if pq.path != nil {
		prev, err := pq.path(ctx)
		if err != nil {
			return err
		}
		pq.sql = prev
	}
	return nil
}

func (pq *ProviderQuery) sqlAll(ctx context.Context, hooks ...queryHook) ([]*Provider, error) {
	var (
		nodes       = []*Provider{}
		withFKs     = pq.withFKs
		_spec       = pq.querySpec()
		loadedTypes = [3]bool{
			pq.withNumber != nil,
			pq.withComment != nil,
			pq.withBusiness != nil,
		}
	)
	if pq.withNumber != nil {
		withFKs = true
	}
	if withFKs {
		_spec.Node.Columns = append(_spec.Node.Columns, provider.ForeignKeys...)
	}
	_spec.ScanValues = func(columns []string) ([]any, error) {
		return (*Provider).scanValues(nil, columns)
	}
	_spec.Assign = func(columns []string, values []any) error {
		node := &Provider{config: pq.config}
		nodes = append(nodes, node)
		node.Edges.loadedTypes = loadedTypes
		return node.assignValues(columns, values)
	}
	for i := range hooks {
		hooks[i](ctx, _spec)
	}
	if err := sqlgraph.QueryNodes(ctx, pq.driver, _spec); err != nil {
		return nil, err
	}
	if len(nodes) == 0 {
		return nodes, nil
	}
	if query := pq.withNumber; query != nil {
		if err := pq.loadNumber(ctx, query, nodes, nil,
			func(n *Provider, e *Number) { n.Edges.Number = e }); err != nil {
			return nil, err
		}
	}
	if query := pq.withComment; query != nil {
		if err := pq.loadComment(ctx, query, nodes,
			func(n *Provider) { n.Edges.Comment = []*Comment{} },
			func(n *Provider, e *Comment) { n.Edges.Comment = append(n.Edges.Comment, e) }); err != nil {
			return nil, err
		}
	}
	if query := pq.withBusiness; query != nil {
		if err := pq.loadBusiness(ctx, query, nodes, nil,
			func(n *Provider, e *Business) { n.Edges.Business = e }); err != nil {
			return nil, err
		}
	}
	return nodes, nil
}

func (pq *ProviderQuery) loadNumber(ctx context.Context, query *NumberQuery, nodes []*Provider, init func(*Provider), assign func(*Provider, *Number)) error {
	ids := make([]int, 0, len(nodes))
	nodeids := make(map[int][]*Provider)
	for i := range nodes {
		if nodes[i].number_provider == nil {
			continue
		}
		fk := *nodes[i].number_provider
		if _, ok := nodeids[fk]; !ok {
			ids = append(ids, fk)
		}
		nodeids[fk] = append(nodeids[fk], nodes[i])
	}
	if len(ids) == 0 {
		return nil
	}
	query.Where(number.IDIn(ids...))
	neighbors, err := query.All(ctx)
	if err != nil {
		return err
	}
	for _, n := range neighbors {
		nodes, ok := nodeids[n.ID]
		if !ok {
			return fmt.Errorf(`unexpected foreign-key "number_provider" returned %v`, n.ID)
		}
		for i := range nodes {
			assign(nodes[i], n)
		}
	}
	return nil
}
func (pq *ProviderQuery) loadComment(ctx context.Context, query *CommentQuery, nodes []*Provider, init func(*Provider), assign func(*Provider, *Comment)) error {
	fks := make([]driver.Value, 0, len(nodes))
	nodeids := make(map[int]*Provider)
	for i := range nodes {
		fks = append(fks, nodes[i].ID)
		nodeids[nodes[i].ID] = nodes[i]
		if init != nil {
			init(nodes[i])
		}
	}
	query.withFKs = true
	query.Where(predicate.Comment(func(s *sql.Selector) {
		s.Where(sql.InValues(s.C(provider.CommentColumn), fks...))
	}))
	neighbors, err := query.All(ctx)
	if err != nil {
		return err
	}
	for _, n := range neighbors {
		fk := n.provider_comment
		if fk == nil {
			return fmt.Errorf(`foreign-key "provider_comment" is nil for node %v`, n.ID)
		}
		node, ok := nodeids[*fk]
		if !ok {
			return fmt.Errorf(`unexpected referenced foreign-key "provider_comment" returned %v for node %v`, *fk, n.ID)
		}
		assign(node, n)
	}
	return nil
}
func (pq *ProviderQuery) loadBusiness(ctx context.Context, query *BusinessQuery, nodes []*Provider, init func(*Provider), assign func(*Provider, *Business)) error {
	fks := make([]driver.Value, 0, len(nodes))
	nodeids := make(map[int]*Provider)
	for i := range nodes {
		fks = append(fks, nodes[i].ID)
		nodeids[nodes[i].ID] = nodes[i]
	}
	query.withFKs = true
	query.Where(predicate.Business(func(s *sql.Selector) {
		s.Where(sql.InValues(s.C(provider.BusinessColumn), fks...))
	}))
	neighbors, err := query.All(ctx)
	if err != nil {
		return err
	}
	for _, n := range neighbors {
		fk := n.provider_business
		if fk == nil {
			return fmt.Errorf(`foreign-key "provider_business" is nil for node %v`, n.ID)
		}
		node, ok := nodeids[*fk]
		if !ok {
			return fmt.Errorf(`unexpected referenced foreign-key "provider_business" returned %v for node %v`, *fk, n.ID)
		}
		assign(node, n)
	}
	return nil
}

func (pq *ProviderQuery) sqlCount(ctx context.Context) (int, error) {
	_spec := pq.querySpec()
	_spec.Node.Columns = pq.ctx.Fields
	if len(pq.ctx.Fields) > 0 {
		_spec.Unique = pq.ctx.Unique != nil && *pq.ctx.Unique
	}
	return sqlgraph.CountNodes(ctx, pq.driver, _spec)
}

func (pq *ProviderQuery) querySpec() *sqlgraph.QuerySpec {
	_spec := sqlgraph.NewQuerySpec(provider.Table, provider.Columns, sqlgraph.NewFieldSpec(provider.FieldID, field.TypeInt))
	_spec.From = pq.sql
	if unique := pq.ctx.Unique; unique != nil {
		_spec.Unique = *unique
	} else if pq.path != nil {
		_spec.Unique = true
	}
	if fields := pq.ctx.Fields; len(fields) > 0 {
		_spec.Node.Columns = make([]string, 0, len(fields))
		_spec.Node.Columns = append(_spec.Node.Columns, provider.FieldID)
		for i := range fields {
			if fields[i] != provider.FieldID {
				_spec.Node.Columns = append(_spec.Node.Columns, fields[i])
			}
		}
	}
	if ps := pq.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if limit := pq.ctx.Limit; limit != nil {
		_spec.Limit = *limit
	}
	if offset := pq.ctx.Offset; offset != nil {
		_spec.Offset = *offset
	}
	if ps := pq.order; len(ps) > 0 {
		_spec.Order = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	return _spec
}

func (pq *ProviderQuery) sqlQuery(ctx context.Context) *sql.Selector {
	builder := sql.Dialect(pq.driver.Dialect())
	t1 := builder.Table(provider.Table)
	columns := pq.ctx.Fields
	if len(columns) == 0 {
		columns = provider.Columns
	}
	selector := builder.Select(t1.Columns(columns...)...).From(t1)
	if pq.sql != nil {
		selector = pq.sql
		selector.Select(selector.Columns(columns...)...)
	}
	if pq.ctx.Unique != nil && *pq.ctx.Unique {
		selector.Distinct()
	}
	for _, p := range pq.predicates {
		p(selector)
	}
	for _, p := range pq.order {
		p(selector)
	}
	if offset := pq.ctx.Offset; offset != nil {
		// limit is mandatory for offset clause. We start
		// with default value, and override it below if needed.
		selector.Offset(*offset).Limit(math.MaxInt32)
	}
	if limit := pq.ctx.Limit; limit != nil {
		selector.Limit(*limit)
	}
	return selector
}

// ProviderGroupBy is the group-by builder for Provider entities.
type ProviderGroupBy struct {
	selector
	build *ProviderQuery
}

// Aggregate adds the given aggregation functions to the group-by query.
func (pgb *ProviderGroupBy) Aggregate(fns ...AggregateFunc) *ProviderGroupBy {
	pgb.fns = append(pgb.fns, fns...)
	return pgb
}

// Scan applies the selector query and scans the result into the given value.
func (pgb *ProviderGroupBy) Scan(ctx context.Context, v any) error {
	ctx = setContextOp(ctx, pgb.build.ctx, ent.OpQueryGroupBy)
	if err := pgb.build.prepareQuery(ctx); err != nil {
		return err
	}
	return scanWithInterceptors[*ProviderQuery, *ProviderGroupBy](ctx, pgb.build, pgb, pgb.build.inters, v)
}

func (pgb *ProviderGroupBy) sqlScan(ctx context.Context, root *ProviderQuery, v any) error {
	selector := root.sqlQuery(ctx).Select()
	aggregation := make([]string, 0, len(pgb.fns))
	for _, fn := range pgb.fns {
		aggregation = append(aggregation, fn(selector))
	}
	if len(selector.SelectedColumns()) == 0 {
		columns := make([]string, 0, len(*pgb.flds)+len(pgb.fns))
		for _, f := range *pgb.flds {
			columns = append(columns, selector.C(f))
		}
		columns = append(columns, aggregation...)
		selector.Select(columns...)
	}
	selector.GroupBy(selector.Columns(*pgb.flds...)...)
	if err := selector.Err(); err != nil {
		return err
	}
	rows := &sql.Rows{}
	query, args := selector.Query()
	if err := pgb.build.driver.Query(ctx, query, args, rows); err != nil {
		return err
	}
	defer rows.Close()
	return sql.ScanSlice(rows, v)
}

// ProviderSelect is the builder for selecting fields of Provider entities.
type ProviderSelect struct {
	*ProviderQuery
	selector
}

// Aggregate adds the given aggregation functions to the selector query.
func (ps *ProviderSelect) Aggregate(fns ...AggregateFunc) *ProviderSelect {
	ps.fns = append(ps.fns, fns...)
	return ps
}

// Scan applies the selector query and scans the result into the given value.
func (ps *ProviderSelect) Scan(ctx context.Context, v any) error {
	ctx = setContextOp(ctx, ps.ctx, ent.OpQuerySelect)
	if err := ps.prepareQuery(ctx); err != nil {
		return err
	}
	return scanWithInterceptors[*ProviderQuery, *ProviderSelect](ctx, ps.ProviderQuery, ps, ps.inters, v)
}

func (ps *ProviderSelect) sqlScan(ctx context.Context, root *ProviderQuery, v any) error {
	selector := root.sqlQuery(ctx)
	aggregation := make([]string, 0, len(ps.fns))
	for _, fn := range ps.fns {
		aggregation = append(aggregation, fn(selector))
	}
	switch n := len(*ps.selector.flds); {
	case n == 0 && len(aggregation) > 0:
		selector.Select(aggregation...)
	case n != 0 && len(aggregation) > 0:
		selector.AppendSelect(aggregation...)
	}
	rows := &sql.Rows{}
	query, args := selector.Query()
	if err := ps.driver.Query(ctx, query, args, rows); err != nil {
		return err
	}
	defer rows.Close()
	return sql.ScanSlice(rows, v)
}
