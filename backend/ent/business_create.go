// Code generated by ent, DO NOT EDIT.

package ent

import (
	"context"
	"errors"
	"fmt"

	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/schema/field"
	"github.com/Blanco0420/Phone-Number-Check/backend/ent/address"
	"github.com/Blanco0420/Phone-Number-Check/backend/ent/business"
	"github.com/Blanco0420/Phone-Number-Check/backend/ent/provider"
)

// BusinessCreate is the builder for creating a Business entity.
type BusinessCreate struct {
	config
	mutation *BusinessMutation
	hooks    []Hook
}

// SetName sets the "Name" field.
func (bc *BusinessCreate) SetName(s string) *BusinessCreate {
	bc.mutation.SetName(s)
	return bc
}

// SetNillableName sets the "Name" field if the given value is not nil.
func (bc *BusinessCreate) SetNillableName(s *string) *BusinessCreate {
	if s != nil {
		bc.SetName(*s)
	}
	return bc
}

// SetWebsiteLink sets the "WebsiteLink" field.
func (bc *BusinessCreate) SetWebsiteLink(s string) *BusinessCreate {
	bc.mutation.SetWebsiteLink(s)
	return bc
}

// SetNillableWebsiteLink sets the "WebsiteLink" field if the given value is not nil.
func (bc *BusinessCreate) SetNillableWebsiteLink(s *string) *BusinessCreate {
	if s != nil {
		bc.SetWebsiteLink(*s)
	}
	return bc
}

// SetOverview sets the "Overview" field.
func (bc *BusinessCreate) SetOverview(s string) *BusinessCreate {
	bc.mutation.SetOverview(s)
	return bc
}

// SetNillableOverview sets the "Overview" field if the given value is not nil.
func (bc *BusinessCreate) SetNillableOverview(s *string) *BusinessCreate {
	if s != nil {
		bc.SetOverview(*s)
	}
	return bc
}

// SetProviderID sets the "provider" edge to the Provider entity by ID.
func (bc *BusinessCreate) SetProviderID(id int) *BusinessCreate {
	bc.mutation.SetProviderID(id)
	return bc
}

// SetProvider sets the "provider" edge to the Provider entity.
func (bc *BusinessCreate) SetProvider(p *Provider) *BusinessCreate {
	return bc.SetProviderID(p.ID)
}

// AddAddresIDs adds the "address" edge to the Address entity by IDs.
func (bc *BusinessCreate) AddAddresIDs(ids ...int) *BusinessCreate {
	bc.mutation.AddAddresIDs(ids...)
	return bc
}

// AddAddress adds the "address" edges to the Address entity.
func (bc *BusinessCreate) AddAddress(a ...*Address) *BusinessCreate {
	ids := make([]int, len(a))
	for i := range a {
		ids[i] = a[i].ID
	}
	return bc.AddAddresIDs(ids...)
}

// Mutation returns the BusinessMutation object of the builder.
func (bc *BusinessCreate) Mutation() *BusinessMutation {
	return bc.mutation
}

// Save creates the Business in the database.
func (bc *BusinessCreate) Save(ctx context.Context) (*Business, error) {
	return withHooks(ctx, bc.sqlSave, bc.mutation, bc.hooks)
}

// SaveX calls Save and panics if Save returns an error.
func (bc *BusinessCreate) SaveX(ctx context.Context) *Business {
	v, err := bc.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (bc *BusinessCreate) Exec(ctx context.Context) error {
	_, err := bc.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (bc *BusinessCreate) ExecX(ctx context.Context) {
	if err := bc.Exec(ctx); err != nil {
		panic(err)
	}
}

// check runs all checks and user-defined validators on the builder.
func (bc *BusinessCreate) check() error {
	if len(bc.mutation.ProviderIDs()) == 0 {
		return &ValidationError{Name: "provider", err: errors.New(`ent: missing required edge "Business.provider"`)}
	}
	return nil
}

func (bc *BusinessCreate) sqlSave(ctx context.Context) (*Business, error) {
	if err := bc.check(); err != nil {
		return nil, err
	}
	_node, _spec := bc.createSpec()
	if err := sqlgraph.CreateNode(ctx, bc.driver, _spec); err != nil {
		if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return nil, err
	}
	id := _spec.ID.Value.(int64)
	_node.ID = int(id)
	bc.mutation.id = &_node.ID
	bc.mutation.done = true
	return _node, nil
}

func (bc *BusinessCreate) createSpec() (*Business, *sqlgraph.CreateSpec) {
	var (
		_node = &Business{config: bc.config}
		_spec = sqlgraph.NewCreateSpec(business.Table, sqlgraph.NewFieldSpec(business.FieldID, field.TypeInt))
	)
	if value, ok := bc.mutation.Name(); ok {
		_spec.SetField(business.FieldName, field.TypeString, value)
		_node.Name = value
	}
	if value, ok := bc.mutation.WebsiteLink(); ok {
		_spec.SetField(business.FieldWebsiteLink, field.TypeString, value)
		_node.WebsiteLink = value
	}
	if value, ok := bc.mutation.Overview(); ok {
		_spec.SetField(business.FieldOverview, field.TypeString, value)
		_node.Overview = value
	}
	if nodes := bc.mutation.ProviderIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2O,
			Inverse: true,
			Table:   business.ProviderTable,
			Columns: []string{business.ProviderColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(provider.FieldID, field.TypeInt),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_node.provider_business = &nodes[0]
		_spec.Edges = append(_spec.Edges, edge)
	}
	if nodes := bc.mutation.AddressIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: false,
			Table:   business.AddressTable,
			Columns: []string{business.AddressColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(address.FieldID, field.TypeInt),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges = append(_spec.Edges, edge)
	}
	return _node, _spec
}

// BusinessCreateBulk is the builder for creating many Business entities in bulk.
type BusinessCreateBulk struct {
	config
	err      error
	builders []*BusinessCreate
}

// Save creates the Business entities in the database.
func (bcb *BusinessCreateBulk) Save(ctx context.Context) ([]*Business, error) {
	if bcb.err != nil {
		return nil, bcb.err
	}
	specs := make([]*sqlgraph.CreateSpec, len(bcb.builders))
	nodes := make([]*Business, len(bcb.builders))
	mutators := make([]Mutator, len(bcb.builders))
	for i := range bcb.builders {
		func(i int, root context.Context) {
			builder := bcb.builders[i]
			var mut Mutator = MutateFunc(func(ctx context.Context, m Mutation) (Value, error) {
				mutation, ok := m.(*BusinessMutation)
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
					_, err = mutators[i+1].Mutate(root, bcb.builders[i+1].mutation)
				} else {
					spec := &sqlgraph.BatchCreateSpec{Nodes: specs}
					// Invoke the actual operation on the latest mutation in the chain.
					if err = sqlgraph.BatchCreate(ctx, bcb.driver, spec); err != nil {
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
		if _, err := mutators[0].Mutate(ctx, bcb.builders[0].mutation); err != nil {
			return nil, err
		}
	}
	return nodes, nil
}

// SaveX is like Save, but panics if an error occurs.
func (bcb *BusinessCreateBulk) SaveX(ctx context.Context) []*Business {
	v, err := bcb.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (bcb *BusinessCreateBulk) Exec(ctx context.Context) error {
	_, err := bcb.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (bcb *BusinessCreateBulk) ExecX(ctx context.Context) {
	if err := bcb.Exec(ctx); err != nil {
		panic(err)
	}
}
