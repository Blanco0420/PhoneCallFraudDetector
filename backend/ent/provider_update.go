// Code generated by ent, DO NOT EDIT.

package ent

import (
	"context"
	"errors"
	"fmt"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/schema/field"
	"github.com/Blanco0420/Phone-Number-Check/backend/ent/business"
	"github.com/Blanco0420/Phone-Number-Check/backend/ent/comment"
	"github.com/Blanco0420/Phone-Number-Check/backend/ent/number"
	"github.com/Blanco0420/Phone-Number-Check/backend/ent/predicate"
	"github.com/Blanco0420/Phone-Number-Check/backend/ent/provider"
)

// ProviderUpdate is the builder for updating Provider entities.
type ProviderUpdate struct {
	config
	hooks    []Hook
	mutation *ProviderMutation
}

// Where appends a list predicates to the ProviderUpdate builder.
func (pu *ProviderUpdate) Where(ps ...predicate.Provider) *ProviderUpdate {
	pu.mutation.Where(ps...)
	return pu
}

// SetName sets the "Name" field.
func (pu *ProviderUpdate) SetName(s string) *ProviderUpdate {
	pu.mutation.SetName(s)
	return pu
}

// SetNillableName sets the "Name" field if the given value is not nil.
func (pu *ProviderUpdate) SetNillableName(s *string) *ProviderUpdate {
	if s != nil {
		pu.SetName(*s)
	}
	return pu
}

// SetNumberID sets the "number" edge to the Number entity by ID.
func (pu *ProviderUpdate) SetNumberID(id int) *ProviderUpdate {
	pu.mutation.SetNumberID(id)
	return pu
}

// SetNumber sets the "number" edge to the Number entity.
func (pu *ProviderUpdate) SetNumber(n *Number) *ProviderUpdate {
	return pu.SetNumberID(n.ID)
}

// AddCommentIDs adds the "comment" edge to the Comment entity by IDs.
func (pu *ProviderUpdate) AddCommentIDs(ids ...int) *ProviderUpdate {
	pu.mutation.AddCommentIDs(ids...)
	return pu
}

// AddComment adds the "comment" edges to the Comment entity.
func (pu *ProviderUpdate) AddComment(c ...*Comment) *ProviderUpdate {
	ids := make([]int, len(c))
	for i := range c {
		ids[i] = c[i].ID
	}
	return pu.AddCommentIDs(ids...)
}

// SetBusinessID sets the "business" edge to the Business entity by ID.
func (pu *ProviderUpdate) SetBusinessID(id int) *ProviderUpdate {
	pu.mutation.SetBusinessID(id)
	return pu
}

// SetNillableBusinessID sets the "business" edge to the Business entity by ID if the given value is not nil.
func (pu *ProviderUpdate) SetNillableBusinessID(id *int) *ProviderUpdate {
	if id != nil {
		pu = pu.SetBusinessID(*id)
	}
	return pu
}

// SetBusiness sets the "business" edge to the Business entity.
func (pu *ProviderUpdate) SetBusiness(b *Business) *ProviderUpdate {
	return pu.SetBusinessID(b.ID)
}

// Mutation returns the ProviderMutation object of the builder.
func (pu *ProviderUpdate) Mutation() *ProviderMutation {
	return pu.mutation
}

// ClearNumber clears the "number" edge to the Number entity.
func (pu *ProviderUpdate) ClearNumber() *ProviderUpdate {
	pu.mutation.ClearNumber()
	return pu
}

// ClearComment clears all "comment" edges to the Comment entity.
func (pu *ProviderUpdate) ClearComment() *ProviderUpdate {
	pu.mutation.ClearComment()
	return pu
}

// RemoveCommentIDs removes the "comment" edge to Comment entities by IDs.
func (pu *ProviderUpdate) RemoveCommentIDs(ids ...int) *ProviderUpdate {
	pu.mutation.RemoveCommentIDs(ids...)
	return pu
}

// RemoveComment removes "comment" edges to Comment entities.
func (pu *ProviderUpdate) RemoveComment(c ...*Comment) *ProviderUpdate {
	ids := make([]int, len(c))
	for i := range c {
		ids[i] = c[i].ID
	}
	return pu.RemoveCommentIDs(ids...)
}

// ClearBusiness clears the "business" edge to the Business entity.
func (pu *ProviderUpdate) ClearBusiness() *ProviderUpdate {
	pu.mutation.ClearBusiness()
	return pu
}

// Save executes the query and returns the number of nodes affected by the update operation.
func (pu *ProviderUpdate) Save(ctx context.Context) (int, error) {
	return withHooks(ctx, pu.sqlSave, pu.mutation, pu.hooks)
}

// SaveX is like Save, but panics if an error occurs.
func (pu *ProviderUpdate) SaveX(ctx context.Context) int {
	affected, err := pu.Save(ctx)
	if err != nil {
		panic(err)
	}
	return affected
}

// Exec executes the query.
func (pu *ProviderUpdate) Exec(ctx context.Context) error {
	_, err := pu.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (pu *ProviderUpdate) ExecX(ctx context.Context) {
	if err := pu.Exec(ctx); err != nil {
		panic(err)
	}
}

// check runs all checks and user-defined validators on the builder.
func (pu *ProviderUpdate) check() error {
	if pu.mutation.NumberCleared() && len(pu.mutation.NumberIDs()) > 0 {
		return errors.New(`ent: clearing a required unique edge "Provider.number"`)
	}
	return nil
}

func (pu *ProviderUpdate) sqlSave(ctx context.Context) (n int, err error) {
	if err := pu.check(); err != nil {
		return n, err
	}
	_spec := sqlgraph.NewUpdateSpec(provider.Table, provider.Columns, sqlgraph.NewFieldSpec(provider.FieldID, field.TypeInt))
	if ps := pu.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if value, ok := pu.mutation.Name(); ok {
		_spec.SetField(provider.FieldName, field.TypeString, value)
	}
	if pu.mutation.NumberCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: true,
			Table:   provider.NumberTable,
			Columns: []string{provider.NumberColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(number.FieldID, field.TypeInt),
			},
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := pu.mutation.NumberIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: true,
			Table:   provider.NumberTable,
			Columns: []string{provider.NumberColumn},
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
	if pu.mutation.CommentCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: false,
			Table:   provider.CommentTable,
			Columns: []string{provider.CommentColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(comment.FieldID, field.TypeInt),
			},
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := pu.mutation.RemovedCommentIDs(); len(nodes) > 0 && !pu.mutation.CommentCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: false,
			Table:   provider.CommentTable,
			Columns: []string{provider.CommentColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(comment.FieldID, field.TypeInt),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := pu.mutation.CommentIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: false,
			Table:   provider.CommentTable,
			Columns: []string{provider.CommentColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(comment.FieldID, field.TypeInt),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Add = append(_spec.Edges.Add, edge)
	}
	if pu.mutation.BusinessCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2O,
			Inverse: false,
			Table:   provider.BusinessTable,
			Columns: []string{provider.BusinessColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(business.FieldID, field.TypeInt),
			},
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := pu.mutation.BusinessIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2O,
			Inverse: false,
			Table:   provider.BusinessTable,
			Columns: []string{provider.BusinessColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(business.FieldID, field.TypeInt),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Add = append(_spec.Edges.Add, edge)
	}
	if n, err = sqlgraph.UpdateNodes(ctx, pu.driver, _spec); err != nil {
		if _, ok := err.(*sqlgraph.NotFoundError); ok {
			err = &NotFoundError{provider.Label}
		} else if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return 0, err
	}
	pu.mutation.done = true
	return n, nil
}

// ProviderUpdateOne is the builder for updating a single Provider entity.
type ProviderUpdateOne struct {
	config
	fields   []string
	hooks    []Hook
	mutation *ProviderMutation
}

// SetName sets the "Name" field.
func (puo *ProviderUpdateOne) SetName(s string) *ProviderUpdateOne {
	puo.mutation.SetName(s)
	return puo
}

// SetNillableName sets the "Name" field if the given value is not nil.
func (puo *ProviderUpdateOne) SetNillableName(s *string) *ProviderUpdateOne {
	if s != nil {
		puo.SetName(*s)
	}
	return puo
}

// SetNumberID sets the "number" edge to the Number entity by ID.
func (puo *ProviderUpdateOne) SetNumberID(id int) *ProviderUpdateOne {
	puo.mutation.SetNumberID(id)
	return puo
}

// SetNumber sets the "number" edge to the Number entity.
func (puo *ProviderUpdateOne) SetNumber(n *Number) *ProviderUpdateOne {
	return puo.SetNumberID(n.ID)
}

// AddCommentIDs adds the "comment" edge to the Comment entity by IDs.
func (puo *ProviderUpdateOne) AddCommentIDs(ids ...int) *ProviderUpdateOne {
	puo.mutation.AddCommentIDs(ids...)
	return puo
}

// AddComment adds the "comment" edges to the Comment entity.
func (puo *ProviderUpdateOne) AddComment(c ...*Comment) *ProviderUpdateOne {
	ids := make([]int, len(c))
	for i := range c {
		ids[i] = c[i].ID
	}
	return puo.AddCommentIDs(ids...)
}

// SetBusinessID sets the "business" edge to the Business entity by ID.
func (puo *ProviderUpdateOne) SetBusinessID(id int) *ProviderUpdateOne {
	puo.mutation.SetBusinessID(id)
	return puo
}

// SetNillableBusinessID sets the "business" edge to the Business entity by ID if the given value is not nil.
func (puo *ProviderUpdateOne) SetNillableBusinessID(id *int) *ProviderUpdateOne {
	if id != nil {
		puo = puo.SetBusinessID(*id)
	}
	return puo
}

// SetBusiness sets the "business" edge to the Business entity.
func (puo *ProviderUpdateOne) SetBusiness(b *Business) *ProviderUpdateOne {
	return puo.SetBusinessID(b.ID)
}

// Mutation returns the ProviderMutation object of the builder.
func (puo *ProviderUpdateOne) Mutation() *ProviderMutation {
	return puo.mutation
}

// ClearNumber clears the "number" edge to the Number entity.
func (puo *ProviderUpdateOne) ClearNumber() *ProviderUpdateOne {
	puo.mutation.ClearNumber()
	return puo
}

// ClearComment clears all "comment" edges to the Comment entity.
func (puo *ProviderUpdateOne) ClearComment() *ProviderUpdateOne {
	puo.mutation.ClearComment()
	return puo
}

// RemoveCommentIDs removes the "comment" edge to Comment entities by IDs.
func (puo *ProviderUpdateOne) RemoveCommentIDs(ids ...int) *ProviderUpdateOne {
	puo.mutation.RemoveCommentIDs(ids...)
	return puo
}

// RemoveComment removes "comment" edges to Comment entities.
func (puo *ProviderUpdateOne) RemoveComment(c ...*Comment) *ProviderUpdateOne {
	ids := make([]int, len(c))
	for i := range c {
		ids[i] = c[i].ID
	}
	return puo.RemoveCommentIDs(ids...)
}

// ClearBusiness clears the "business" edge to the Business entity.
func (puo *ProviderUpdateOne) ClearBusiness() *ProviderUpdateOne {
	puo.mutation.ClearBusiness()
	return puo
}

// Where appends a list predicates to the ProviderUpdate builder.
func (puo *ProviderUpdateOne) Where(ps ...predicate.Provider) *ProviderUpdateOne {
	puo.mutation.Where(ps...)
	return puo
}

// Select allows selecting one or more fields (columns) of the returned entity.
// The default is selecting all fields defined in the entity schema.
func (puo *ProviderUpdateOne) Select(field string, fields ...string) *ProviderUpdateOne {
	puo.fields = append([]string{field}, fields...)
	return puo
}

// Save executes the query and returns the updated Provider entity.
func (puo *ProviderUpdateOne) Save(ctx context.Context) (*Provider, error) {
	return withHooks(ctx, puo.sqlSave, puo.mutation, puo.hooks)
}

// SaveX is like Save, but panics if an error occurs.
func (puo *ProviderUpdateOne) SaveX(ctx context.Context) *Provider {
	node, err := puo.Save(ctx)
	if err != nil {
		panic(err)
	}
	return node
}

// Exec executes the query on the entity.
func (puo *ProviderUpdateOne) Exec(ctx context.Context) error {
	_, err := puo.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (puo *ProviderUpdateOne) ExecX(ctx context.Context) {
	if err := puo.Exec(ctx); err != nil {
		panic(err)
	}
}

// check runs all checks and user-defined validators on the builder.
func (puo *ProviderUpdateOne) check() error {
	if puo.mutation.NumberCleared() && len(puo.mutation.NumberIDs()) > 0 {
		return errors.New(`ent: clearing a required unique edge "Provider.number"`)
	}
	return nil
}

func (puo *ProviderUpdateOne) sqlSave(ctx context.Context) (_node *Provider, err error) {
	if err := puo.check(); err != nil {
		return _node, err
	}
	_spec := sqlgraph.NewUpdateSpec(provider.Table, provider.Columns, sqlgraph.NewFieldSpec(provider.FieldID, field.TypeInt))
	id, ok := puo.mutation.ID()
	if !ok {
		return nil, &ValidationError{Name: "id", err: errors.New(`ent: missing "Provider.id" for update`)}
	}
	_spec.Node.ID.Value = id
	if fields := puo.fields; len(fields) > 0 {
		_spec.Node.Columns = make([]string, 0, len(fields))
		_spec.Node.Columns = append(_spec.Node.Columns, provider.FieldID)
		for _, f := range fields {
			if !provider.ValidColumn(f) {
				return nil, &ValidationError{Name: f, err: fmt.Errorf("ent: invalid field %q for query", f)}
			}
			if f != provider.FieldID {
				_spec.Node.Columns = append(_spec.Node.Columns, f)
			}
		}
	}
	if ps := puo.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if value, ok := puo.mutation.Name(); ok {
		_spec.SetField(provider.FieldName, field.TypeString, value)
	}
	if puo.mutation.NumberCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: true,
			Table:   provider.NumberTable,
			Columns: []string{provider.NumberColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(number.FieldID, field.TypeInt),
			},
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := puo.mutation.NumberIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: true,
			Table:   provider.NumberTable,
			Columns: []string{provider.NumberColumn},
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
	if puo.mutation.CommentCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: false,
			Table:   provider.CommentTable,
			Columns: []string{provider.CommentColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(comment.FieldID, field.TypeInt),
			},
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := puo.mutation.RemovedCommentIDs(); len(nodes) > 0 && !puo.mutation.CommentCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: false,
			Table:   provider.CommentTable,
			Columns: []string{provider.CommentColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(comment.FieldID, field.TypeInt),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := puo.mutation.CommentIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: false,
			Table:   provider.CommentTable,
			Columns: []string{provider.CommentColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(comment.FieldID, field.TypeInt),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Add = append(_spec.Edges.Add, edge)
	}
	if puo.mutation.BusinessCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2O,
			Inverse: false,
			Table:   provider.BusinessTable,
			Columns: []string{provider.BusinessColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(business.FieldID, field.TypeInt),
			},
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := puo.mutation.BusinessIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2O,
			Inverse: false,
			Table:   provider.BusinessTable,
			Columns: []string{provider.BusinessColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(business.FieldID, field.TypeInt),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Add = append(_spec.Edges.Add, edge)
	}
	_node = &Provider{config: puo.config}
	_spec.Assign = _node.assignValues
	_spec.ScanValues = _node.scanValues
	if err = sqlgraph.UpdateNode(ctx, puo.driver, _spec); err != nil {
		if _, ok := err.(*sqlgraph.NotFoundError); ok {
			err = &NotFoundError{provider.Label}
		} else if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return nil, err
	}
	puo.mutation.done = true
	return _node, nil
}
