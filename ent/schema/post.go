package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/moroz/uuidv7-go"
	"time"
)

// Post holds the schema definition for the Post entity.
type Post struct {
	ent.Schema
}

// Fields of the Post.
func (Post) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuidv7.UUID{}).Default(uuidv7.Generate).Immutable().Unique(),
		field.Enum("type").Values(string(MediaTypePhoto), string(MediaTypeVideo), string(MediaTypeAnimation)),
		field.String("file_id"),
		field.Bool("is_sent").Default(false),
		field.Time("created_at").Default(time.Now),
		field.Time("sent_at").Optional(),
	}
}

// Edges of the Post.
func (Post) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("message_ids", PostMessageId.Type),
		edge.To("image_hash", ImageHash.Type).Unique(),
	}
}

func (Post) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("is_sent"),
		index.Fields("type"),
	}
}
