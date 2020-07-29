package gqlserver

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/graphql-go/graphql"
)

var (
	nodeTypes = make(map[string]*NodeType)

	errNoRetrieve = errors.New("cannot retrieve Node of this type")
	errNoDelete   = errors.New("cannot delete Node of this type")
)

// NodeType defines a Node subtype.
type NodeType struct {
	prefix string
	typ    reflect.Type

	object *graphql.Object

	// GetID extracts ID (without prefix) from the source object.
	GetID func(source interface{}) string

	// Retrieve fetches an object from ID (without prefix).
	Retrieve func(id string) (interface{}, error)

	// Delete destroys an object.
	Delete func(source interface{}) error
}

// NewNodeType creates a NodeType.
func NewNodeType(value interface{}) *NodeType {
	return NewNodeTypeNamed("", value)
}

// NewNodeTypeNamed creates a NodeType with specified name.
func NewNodeTypeNamed(name string, value interface{}) (nt *NodeType) {
	typ := reflect.TypeOf(value)
	if typ.Kind() == reflect.Ptr {
		if elem := typ.Elem(); elem.Kind() == reflect.Interface {
			typ = elem
		}
	}
	if name == "" {
		name = strings.ToLower(typ.Name())
	}

	nt = &NodeType{
		prefix: name,
		typ:    typ,
	}
	return nt
}

// Annotate updates ObjectConfig with Node interface and "id" field.
//
// The 'id' can be resolved from:
//  - nt.GetID function.
//  - ObjectConfig 'nid' field of NonNullInt or NonNullString type.
// If neither is present, this function panics.
func (nt *NodeType) Annotate(oc graphql.ObjectConfig) graphql.ObjectConfig {
	if oc.Interfaces == nil {
		oc.Interfaces = []*graphql.Interface{}
	}
	oc.Interfaces = append(oc.Interfaces.([]*graphql.Interface), nodeInterface)

	if oc.Fields == nil {
		oc.Fields = graphql.Fields{}
	}
	fields := oc.Fields.(graphql.Fields)
	nidField := fields["nid"]

	var resolve graphql.FieldResolveFn
	switch {
	case nt.GetID != nil:
		resolve = func(p graphql.ResolveParams) (interface{}, error) {
			return fmt.Sprintf("%s:%s", nt.prefix, nt.GetID(p.Source)), nil
		}
	case nidField != nil:
		switch nidField.Type {
		case NonNullID, NonNullInt, NonNullString:
			resolve = func(p graphql.ResolveParams) (interface{}, error) {
				nid, e := nidField.Resolve(p)
				if e != nil {
					return nil, e
				}
				return fmt.Sprintf("%s:%v", nt.prefix, nid), nil
			}
		}
	}
	if resolve == nil {
		panic("cannot resolve 'id' field")
	}

	fields["id"] = &graphql.Field{
		Type:        graphql.NewNonNull(graphql.ID),
		Description: "Globally unique ID.",
		Resolve:     resolve,
	}

	return oc
}

// Register enables accessing Node of this type by ID.
func (nt *NodeType) Register(object *graphql.Object) {
	nt.object = object
	Schema.Types = append(Schema.Types, object)

	if nodeTypes[nt.prefix] != nil {
		panic("duplicate prefix " + nt.prefix)
	}
	nodeTypes[nt.prefix] = nt
}

var nodeInterface = graphql.NewInterface(graphql.InterfaceConfig{
	Name: "Node",
	Fields: graphql.Fields{
		"id": &graphql.Field{
			Type: NonNullID,
		},
	},
	ResolveType: func(p graphql.ResolveTypeParams) *graphql.Object {
		typ := reflect.TypeOf(p.Value)
		for _, nt := range nodeTypes {
			if typ.AssignableTo(nt.typ) {
				return nt.object
			}
		}
		return nil
	},
})

func initNode() {
	AddQuery(&graphql.Field{
		Name:        "node",
		Description: "Retrieve object by global ID.",
		Type:        nodeInterface,
		Args: graphql.FieldConfigArgument{
			"id": &graphql.ArgumentConfig{
				Type: NonNullID,
			},
		},
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			id := p.Args["id"].(string)
			tokens := strings.SplitN(id, ":", 2)
			nt := nodeTypes[tokens[0]]
			if nt == nil || nt.Retrieve == nil {
				return nil, errNoRetrieve
			}
			return nt.Retrieve(tokens[1])
		},
	})

	AddMutation(&graphql.Field{
		Name:        "delete",
		Description: "Delete object by global ID. The result indicates whether the object previously exists.",
		Type:        graphql.Boolean,
		Args: graphql.FieldConfigArgument{
			"id": &graphql.ArgumentConfig{
				Type: NonNullID,
			},
		},
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			id := p.Args["id"].(string)
			tokens := strings.SplitN(id, ":", 2)
			nt := nodeTypes[tokens[0]]
			if nt == nil || nt.Retrieve == nil {
				return nil, errNoRetrieve
			}

			obj, e := nt.Retrieve(tokens[1])
			if e != nil {
				return nil, e
			}
			if obj == nil {
				return false, nil
			}

			if nt.Delete == nil {
				return true, errNoDelete
			}
			return true, nt.Delete(obj)
		},
	})
}
