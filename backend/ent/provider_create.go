// Code generated by ent, DO NOT EDIT.

package ent

import (
	"context"
	"errors"
	"fmt"

	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/schema/field"
	"github.com/Blanco0420/Phone-Number-Check/backend/ent/business"
	"github.com/Blanco0420/Phone-Number-Check/backend/ent/comment"
	"github.com/Blanco0420/Phone-Number-Check/backend/ent/number"
	"github.com/Blanco0420/Phone-Number-Check/backend/ent/provider"
)

// ProviderCreate is the builder for creating a Provider entity.
type ProviderCreate struct {
	config
	mutation *ProviderMutation
	hooks    []Hook
}

// SetName sets the "Name" field.
func (pc *ProviderCreate) SetName(s string) *ProviderCreate {
	pc.mutation.SetName(s)
	return pc
}

// SetNumberID sets the "number" edge to the Number entity by ID.
func (pc *ProviderCreate) SetNumberID(id int) *ProviderCreate {
	pc.mutation.SetNumberID(id)
	return pc
}

// SetNumber sets the "number" edge to the Number entity.
func (pc *ProviderCreate) SetNumber(n *Number) *ProviderCreate {
	return pc.SetNumberID(n.ID)
}

// AddCommentIDs adds the "comment" edge to the Comment entity by IDs.
func (pc *ProviderCreate) AddCommentIDs(ids ...int) *ProviderCreate {
	pc.mutation.AddCommentIDs(ids...)
	return pc
}

// AddComment adds the "comment" edges to the Comment entity.
func (pc *ProviderCreate) AddComment(c ...*Comment) *ProviderCreate {
	ids := make([]int, len(c))
	for i := range c {
		ids[i] = c[i].ID
	}
	return pc.AddCommentIDs(ids...)
}

// SetBusinessID sets the "business" edge to the Business entity by ID.
func (pc *ProviderCreate) SetBusinessID(id int) *ProviderCreate {
	pc.mutation.SetBusinessID(id)
	return pc
}

// SetNillableBusinessID sets the "business" edge to the Business entity by ID if the given value is not nil.
func (pc *ProviderCreate) SetNillableBusinessID(id *int) *ProviderCreate {
	if id != nil {
		pc = pc.SetBusinessID(*id)
	}
	return pc
}

// SetBusiness sets the "business" edge to the Business entity.
func (pc *ProviderCreate) SetBusiness(b *Business) *ProviderCreate {
	return pc.SetBusinessID(b.ID)
}

// Mutation returns the ProviderMutation object of the builder.
func (pc *ProviderCreate) Mutation() *ProviderMutation {
	return pc.mutation
}

// Save creates the Provider in the database.
func (pc *ProviderCreate) Save(ctx context.Context) (*Provider, error) {
	return withHooks(ctx, pc.sqlSave, pc.mutation, pc.hooks)
}

// SaveX calls Save and panics if Save returns an error.
func (pc *ProviderCreate) SaveX(ctx context.Context) *Provider {
	v, err := pc.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (pc *ProviderCreate) Exec(ctx context.Context) error {
	_, err := pc.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (pc *ProviderCreate) ExecX(ctx context.Context) {
	if err := pc.Exec(ctx); err != nil {
		panic(err)
	}
}

// check runs all checks and user-defined validators on the builder.
func (pc *ProviderCreate) check() error {
	if _, ok := pc.mutation.Name(); !ok {
		return &ValidationError{Name: "Name", err: errors.New(`ent: missing required field "Provider.Name"`)}
	}
	if len(pc.mutation.NumberIDs()) == 0 {
		return &ValidationError{Name: "number", err: errors.New(`ent: missing required edge "Provider.number"`)}
	}
	return nil
}

func (pc *ProviderCreate) sqlSave(ctx context.Context) (*Provider, error) {
	if err := pc.check(); err != nil {
		return nil, err
	}
	_node, _spec := pc.createSpec()
	if err := sqlgraph.CreateNode(ctx, pc.driver, _spec); err != nil {
		if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return nil, err
	}
	id := _spec.ID.Value.(int64)
	_node.ID = int(id)
	pc.mutation.id = &_node.ID
	pc.mutation.done = true
	return _node, nil
}

func (pc *ProviderCreate) createSpec() (*Provider, *sqlgraph.CreateSpec) {
	var (
		_node = &Provider{config: pc.config}
		_spec = sqlgraph.NewCreateSpec(provider.Table, sqlgraph.NewFieldSpec(provider.FieldID, field.TypeInt))
	)
	if value, ok := pc.mutation.Name(); ok {
		_spec.SetField(provider.FieldName, field.TypeString, value)
		_node.Name = value
	}
	if nodes := pc.mutation.NumberIDs(); len(nodes) > 0 {
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
		_node.number_provider = &nodes[0]
		_spec.Edges = append(_spec.Edges, edge)
	}
	if nodes := pc.mutation.CommentIDs(); len(nodes) > 0 {
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
		_spec.Edges = append(_spec.Edges, edge)
	}
	if nodes := pc.mutation.BusinessIDs(); len(nodes) > 0 {
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
		_spec.Edges = append(_spec.Edges, edge)
	}
	return _node, _spec
}

// ProviderCreateBulk is the builder for creating many Provider entities in bulk.
type ProviderCreateBulk struct {
	config
	err      error
	builders []*ProviderCreate
}

// Save creates the Provider entities in the database.
func (pcb *ProviderCreateBulk) Save(ctx context.Context) ([]*Provider, error) {
	if pcb.err != nil {
		return nil, pcb.err
	}
	specs := make([]*sqlgraph.CreateSpec, len(pcb.builders))
	nodes := make([]*Provider, len(pcb.builders))
	mutators := make([]Mutator, len(pcb.builders))
	for i := range pcb.builders {
		func(i int, root context.Context) {
			builder := pcb.builders[i]
			var mut Mutator = MutateFunc(func(ctx context.Context, m Mutation) (Value, error) {
				mutation, ok := m.(*ProviderMutation)
				if !ok {
					return nil, fmt.Errorf("unexpected mutation type %T", m)
				}
				if err := builder.check(); err != nil {
					return nil, err
				}
				builder.mutation = mutation
				var err error
				nodes[i], specs[i] = builder.createSpec()
				if i < len(mutators)-1 {
					_, err = mutators[i+1].Mutate(root, pcb.builders[i+1].mutation)
				} else {
					spec := &sqlgraph.BatchCreateSpec{Nodes: specs}
					// Invoke the actual operation on the latest mutation in the chain.
					if err = sqlgraph.BatchCreate(ctx, pcb.driver, spec); err != nil {
						if sqlgraph.IsConstraintError(err) {
							err = &ConstraintError{msg: err.Error(), wrap: err}
						}
					}
				}
				if err != nil {
					return nil, err
				}
				mutation.id = &nodes[i].ID
				if specs[i].ID.Value != nil {
					id := specs[i].ID.Value.(int64)
					nodes[i].ID = int(id)
				}
				mutation.done = true
				return nodes[i], nil
			})
			for i := len(builder.hooks) - 1; i >= 0; i-- {
				mut = builder.hooks[i](mut)
			}
			mutators[i] = mut
		}(i, ctx)
	}
	if len(mutators) > 0 {
		if _, err := mutators[0].Mutate(ctx, pcb.builders[0].mutation); err != nil {
			return nil, err
		}
	}
	return nodes, nil
}

// SaveX is like Save, but panics if an error occurs.
func (pcb *ProviderCreateBulk) SaveX(ctx context.Context) []*Provider {
	v, err := pcb.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (pcb *ProviderCreateBulk) Exec(ctx context.Context) error {
	_, err := pcb.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (pcb *ProviderCreateBulk) ExecX(ctx context.Context) {
	if err := pcb.Exec(ctx); err != nil {
		panic(err)
	}
}
