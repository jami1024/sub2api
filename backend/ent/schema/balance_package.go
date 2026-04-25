package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// BalancePackage holds the schema definition for the BalancePackage entity.
//
// 删除策略：硬删除
// BalancePackage 作为商品配置项，通过 for_sale 控制上下架；
// 若已有历史订单引用，外键约束会阻止物理删除。
type BalancePackage struct {
	ent.Schema
}

func (BalancePackage) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "balance_packages"},
	}
}

func (BalancePackage) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").
			MaxLen(100).
			NotEmpty(),
		field.String("description").
			SchemaType(map[string]string{dialect.Postgres: "text"}).
			Default(""),
		field.Float("price").
			SchemaType(map[string]string{dialect.Postgres: "decimal(20,2)"}),
		field.Float("credit_amount").
			SchemaType(map[string]string{dialect.Postgres: "decimal(20,8)"}),
		field.String("package_scope").
			MaxLen(20).
			NotEmpty(),
		field.String("product_name").
			MaxLen(100).
			Default(""),
		field.Bool("for_sale").
			Default(true),
		field.Int("sort_order").
			Default(0),
		field.Time("created_at").
			Immutable().
			Default(time.Now).
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now).
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
	}
}

func (BalancePackage) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("package_scope", "for_sale"),
		index.Fields("sort_order"),
	}
}
