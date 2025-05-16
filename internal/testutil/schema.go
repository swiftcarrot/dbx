package testutil

import "github.com/swiftcarrot/dbx/schema"

func CreateUsersTableSchema() *schema.Schema {
	s := schema.NewSchema()
	s.CreateTable("users", func(t *schema.Table) {
		t.String("name")
		t.Text("bio")
		t.Integer("age")
		t.BigInt("credit")
		t.Float("weight")
		t.Decimal("balance")
		t.DateTime("created_at")
		t.Time("time")
		t.Date("birthday")
		t.Binary("bin")
		t.Boolean("verified")
	})
	return s
}
