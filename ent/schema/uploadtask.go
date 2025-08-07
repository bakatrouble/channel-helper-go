package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/moroz/uuidv7-go"
	"time"
)

// UploadTask holds the schema definition for the UploadTask entity.
type UploadTask struct {
	ent.Schema
}

// Fields of the UploadTask.
func (UploadTask) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuidv7.UUID{}).Default(uuidv7.Generate).Immutable().Unique(),
		field.Enum("type").Values(string(MediaTypePhoto), string(MediaTypeVideo), string(MediaTypeAnimation)),
		field.Bytes("data").Optional(),
		field.Bool("is_processed").Default(false),
		field.Time("created_at").Default(time.Now),
		field.Time("sent_at").Optional(),
		field.String("image_hash").Optional(),
	}
}

// Edges of the UploadTask.
func (UploadTask) Edges() []ent.Edge {
	return nil
}

func (UploadTask) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("image_hash"),
		index.Fields("is_processed"),
		index.Fields("type"),
	}
}
