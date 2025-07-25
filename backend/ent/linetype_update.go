// Code generated by ent, DO NOT EDIT.

package ent

import (
	"context"
	"errors"
	"fmt"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/schema/field"
	"github.com/Blanco0420/Phone-Number-Check/backend/ent/linetype"
	"github.com/Blanco0420/Phone-Number-Check/backend/ent/number"
	"github.com/Blanco0420/Phone-Number-Check/backend/ent/predicate"
	"github.com/Blanco0420/Phone-Number-Check/backend/providers"
)

// LineTypeUpdate is the builder for updating LineType entities.
type LineTypeUpdate struct {
	config
	hooks    []Hook
	mutation *LineTypeMutation
}

// Where appends a list predicates to the LineTypeUpdate builder.
func (ltu *LineTypeUpdate) Where(ps ...predicate.LineType) *LineTypeUpdate {
	ltu.mutation.Where(ps...)
	return ltu
}

// SetLineType sets the "LineType" field.
func (ltu *LineTypeUpdate) SetLineType(pt providers.LineType) *LineTypeUpdate {
	ltu.mutation.SetLineType(pt)
	return ltu
}

// SetNillableLineType sets the "LineType" field if the given value is not nil.
func (ltu *LineTypeUpdate) SetNillableLineType(pt *providers.LineType) *LineTypeUpdate {
	if pt != nil {
		ltu.SetLineType(*pt)
	}
	return ltu
}

// AddNumberIDs adds the "number" edge to the Number entity by IDs.
func (ltu *LineTypeUpdate) AddNumberIDs(ids ...int) *LineTypeUpdate {
	ltu.mutation.AddNumberIDs(ids...)
	return ltu
}

// AddNumber adds the "number" edges to the Number entity.
func (ltu *LineTypeUpdate) AddNumber(n ...*Number) *LineTypeUpdate {
	ids := make([]int, len(n))
	for i := range n {
		ids[i] = n[i].ID
	}
	return ltu.AddNumberIDs(ids...)
}

// Mutation returns the LineTypeMutation object of the builder.
func (ltu *LineTypeUpdate) Mutation() *LineTypeMutation {
	return ltu.mutation
}

// ClearNumber clears all "number" edges to the Number entity.
func (ltu *LineTypeUpdate) ClearNumber() *LineTypeUpdate {
	ltu.mutation.ClearNumber()
	return ltu
}

// RemoveNumberIDs removes the "number" edge to Number entities by IDs.
func (ltu *LineTypeUpdate) RemoveNumberIDs(ids ...int) *LineTypeUpdate {
	ltu.mutation.RemoveNumberIDs(ids...)
	return ltu
}

// RemoveNumber removes "number" edges to Number entities.
func (ltu *LineTypeUpdate) RemoveNumber(n ...*Number) *LineTypeUpdate {
	ids := make([]int, len(n))
	for i := range n {
		ids[i] = n[i].ID
	}
	return ltu.RemoveNumberIDs(ids...)
}

// Save executes the query and returns the number of nodes affected by the update operation.
func (ltu *LineTypeUpdate) Save(ctx context.Context) (int, error) {
	return withHooks(ctx, ltu.sqlSave, ltu.mutation, ltu.hooks)
}

// SaveX is like Save, but panics if an error occurs.
func (ltu *LineTypeUpdate) SaveX(ctx context.Context) int {
	affected, err := ltu.Save(ctx)
	if err != nil {
		panic(err)
	}
	return affected
}

// Exec executes the query.
func (ltu *LineTypeUpdate) Exec(ctx context.Context) error {
	_, err := ltu.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (ltu *LineTypeUpdate) ExecX(ctx context.Context) {
	if err := ltu.Exec(ctx); err != nil {
		panic(err)
	}
}

// check runs all checks and user-defined validators on the builder.
func (ltu *LineTypeUpdate) check() error {
	if v, ok := ltu.mutation.LineType(); ok {
		if err := linetype.LineTypeValidator(v); err != nil {
			return &ValidationError{Name: "LineType", err: fmt.Errorf(`ent: validator failed for field "LineType.LineType": %w`, err)}
		}
	}
	return nil
}

func (ltu *LineTypeUpdate) sqlSave(ctx context.Context) (n int, err error) {
	if err := ltu.check(); err != nil {
		return n, err
	}
	_spec := sqlgraph.NewUpdateSpec(linetype.Table, linetype.Columns, sqlgraph.NewFieldSpec(linetype.FieldID, field.TypeInt))
	if ps := ltu.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if value, ok := ltu.mutation.LineType(); ok {
		_spec.SetField(linetype.FieldLineType, field.TypeEnum, value)
	}
	if ltu.mutation.NumberCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: false,
			Table:   linetype.NumberTable,
			Columns: []string{linetype.NumberColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(number.FieldID, field.TypeInt),
			},
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := ltu.mutation.RemovedNumberIDs(); len(nodes) > 0 && !ltu.mutation.NumberCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: false,
			Table:   linetype.NumberTable,
			Columns: []string{linetype.NumberColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(number.FieldID, field.TypeInt),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := ltu.mutation.NumberIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: false,
			Table:   linetype.NumberTable,
			Columns: []string{linetype.NumberColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(number.FieldID, field.TypeInt),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Add = append(_spec.Edges.Add, edge)
	}
	if n, err = sqlgraph.UpdateNodes(ctx, ltu.driver, _spec); err != nil {
		if _, ok := err.(*sqlgraph.NotFoundError); ok {
			err = &NotFoundError{linetype.Label}
		} else if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return 0, err
	}
	ltu.mutation.done = true
	return n, nil
}

// LineTypeUpdateOne is the builder for updating a single LineType entity.
type LineTypeUpdateOne struct {
	config
	fields   []string
	hooks    []Hook
	mutation *LineTypeMutation
}

// SetLineType sets the "LineType" field.
func (ltuo *LineTypeUpdateOne) SetLineType(pt providers.LineType) *LineTypeUpdateOne {
	ltuo.mutation.SetLineType(pt)
	return ltuo
}

// SetNillableLineType sets the "LineType" field if the given value is not nil.
func (ltuo *LineTypeUpdateOne) SetNillableLineType(pt *providers.LineType) *LineTypeUpdateOne {
	if pt != nil {
		ltuo.SetLineType(*pt)
	}
	return ltuo
}

// AddNumberIDs adds the "number" edge to the Number entity by IDs.
func (ltuo *LineTypeUpdateOne) AddNumberIDs(ids ...int) *LineTypeUpdateOne {
	ltuo.mutation.AddNumberIDs(ids...)
	return ltuo
}

// AddNumber adds the "number" edges to the Number entity.
func (ltuo *LineTypeUpdateOne) AddNumber(n ...*Number) *LineTypeUpdateOne {
	ids := make([]int, len(n))
	for i := range n {
		ids[i] = n[i].ID
	}
	return ltuo.AddNumberIDs(ids...)
}

// Mutation returns the LineTypeMutation object of the builder.
func (ltuo *LineTypeUpdateOne) Mutation() *LineTypeMutation {
	return ltuo.mutation
}

// ClearNumber clears all "number" edges to the Number entity.
func (ltuo *LineTypeUpdateOne) ClearNumber() *LineTypeUpdateOne {
	ltuo.mutation.ClearNumber()
	return ltuo
}

// RemoveNumberIDs removes the "number" edge to Number entities by IDs.
func (ltuo *LineTypeUpdateOne) RemoveNumberIDs(ids ...int) *LineTypeUpdateOne {
	ltuo.mutation.RemoveNumberIDs(ids...)
	return ltuo
}

// RemoveNumber removes "number" edges to Number entities.
func (ltuo *LineTypeUpdateOne) RemoveNumber(n ...*Number) *LineTypeUpdateOne {
	ids := make([]int, len(n))
	for i := range n {
		ids[i] = n[i].ID
	}
	return ltuo.RemoveNumberIDs(ids...)
}

// Where appends a list predicates to the LineTypeUpdate builder.
func (ltuo *LineTypeUpdateOne) Where(ps ...predicate.LineType) *LineTypeUpdateOne {
	ltuo.mutation.Where(ps...)
	return ltuo
}

// Select allows selecting one or more fields (columns) of the returned entity.
// The default is selecting all fields defined in the entity schema.
func (ltuo *LineTypeUpdateOne) Select(field string, fields ...string) *LineTypeUpdateOne {
	ltuo.fields = append([]string{field}, fields...)
	return ltuo
}

// Save executes the query and returns the updated LineType entity.
func (ltuo *LineTypeUpdateOne) Save(ctx context.Context) (*LineType, error) {
	return withHooks(ctx, ltuo.sqlSave, ltuo.mutation, ltuo.hooks)
}

// SaveX is like Save, but panics if an error occurs.
func (ltuo *LineTypeUpdateOne) SaveX(ctx context.Context) *LineType {
	node, err := ltuo.Save(ctx)
	if err != nil {
		panic(err)
	}
	return node
}

// Exec executes the query on the entity.
func (ltuo *LineTypeUpdateOne) Exec(ctx context.Context) error {
	_, err := ltuo.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (ltuo *LineTypeUpdateOne) ExecX(ctx context.Context) {
	if err := ltuo.Exec(ctx); err != nil {
		panic(err)
	}
}

// check runs all checks and user-defined validators on the builder.
func (ltuo *LineTypeUpdateOne) check() error {
	if v, ok := ltuo.mutation.LineType(); ok {
		if err := linetype.LineTypeValidator(v); err != nil {
			return &ValidationError{Name: "LineType", err: fmt.Errorf(`ent: validator failed for field "LineType.LineType": %w`, err)}
		}
	}
	return nil
}

func (ltuo *LineTypeUpdateOne) sqlSave(ctx context.Context) (_node *LineType, err error) {
	if err := ltuo.check(); err != nil {
		return _node, err
	}
	_spec := sqlgraph.NewUpdateSpec(linetype.Table, linetype.Columns, sqlgraph.NewFieldSpec(linetype.FieldID, field.TypeInt))
	id, ok := ltuo.mutation.ID()
	if !ok {
		return nil, &ValidationError{Name: "id", err: errors.New(`ent: missing "LineType.id" for update`)}
	}
	_spec.Node.ID.Value = id
	if fields := ltuo.fields; len(fields) > 0 {
		_spec.Node.Columns = make([]string, 0, len(fields))
		_spec.Node.Columns = append(_spec.Node.Columns, linetype.FieldID)
		for _, f := range fields {
			if !linetype.ValidColumn(f) {
				return nil, &ValidationError{Name: f, err: fmt.Errorf("ent: invalid field %q for query", f)}
			}
			if f != linetype.FieldID {
				_spec.Node.Columns = append(_spec.Node.Columns, f)
			}
		}
	}
	if ps := ltuo.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if value, ok := ltuo.mutation.LineType(); ok {
		_spec.SetField(linetype.FieldLineType, field.TypeEnum, value)
	}
	if ltuo.mutation.NumberCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: false,
			Table:   linetype.NumberTable,
			Columns: []string{linetype.NumberColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(number.FieldID, field.TypeInt),
			},
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := ltuo.mutation.RemovedNumberIDs(); len(nodes) > 0 && !ltuo.mutation.NumberCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: false,
			Table:   linetype.NumberTable,
			Columns: []string{linetype.NumberColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(number.FieldID, field.TypeInt),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := ltuo.mutation.NumberIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: false,
			Table:   linetype.NumberTable,
			Columns: []string{linetype.NumberColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(number.FieldID, field.TypeInt),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Add = append(_spec.Edges.Add, edge)
	}
	_node = &LineType{config: ltuo.config}
	_spec.Assign = _node.assignValues
	_spec.ScanValues = _node.scanValues
	if err = sqlgraph.UpdateNode(ctx, ltuo.driver, _spec); err != nil {
		if _, ok := err.(*sqlgraph.NotFoundError); ok {
			err = &NotFoundError{linetype.Label}
		} else if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return nil, err
	}
	ltuo.mutation.done = true
	return _node, nil
}
