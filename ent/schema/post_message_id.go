package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// PostMessageId holds the schema definition for the PostMessageId entity.
type PostMessageId struct {
	ent.Schema
}

// Fields of the PostMessageId.
func (PostMessageId) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("chat_id"),
		field.Int("message_id"),
	}
}

// Edges of the PostMessageId.
func (PostMessageId) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("post", Post.Type).
			Ref("message_ids").
			Unique(),
	}
}

func (PostMessageId) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("chat_id", "message_id").
			Unique(),
	}
}
