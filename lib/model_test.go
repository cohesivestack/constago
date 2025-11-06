package constago

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAddStruct_ImportAliasCollisions(t *testing.T) {
	t.Run("no_collision", func(t *testing.T) {
		model := NewModel(nil)

		// Add first struct with import
		struct1 := &StructModel{
			Name: "User",
			Getters: []*GetterOutput{
				{
					Name: "GetName",
					Returns: []*ReturnOutput{
						{
							Value: &ValueOutput{
								TypePackage: &TypePackageOutput{
									Path:  "github.com/example/strings",
									Name:  "strings",
									Alias: "",
								},
							},
						},
					},
				},
			},
		}
		model.AddStruct("github.com/test/package1", "package1", struct1)

		// Add first struct with import
		struct2 := &StructModel{
			Name: "User",
			Getters: []*GetterOutput{
				{
					Name: "GetName",
					Returns: []*ReturnOutput{
						{
							Value: &ValueOutput{
								TypePackage: &TypePackageOutput{
									Path:  "github.com/example/stringss",
									Name:  "stringss",
									Alias: "",
								},
							},
						},
					},
				},
			},
		}
		model.AddStruct("github.com/test/package1", "package1", struct2)

		// Get the package
		pkg := model.Packages["github.com/test/package1"]
		assert.NotNil(t, pkg, "Expected package to exist")

		// Check that the import exists with correct alias
		importPkg, exists := pkg.Imports["github.com/example/strings"]
		assert.True(t, exists, "Expected import to exist")

		assert.Equal(t, "strings", importPkg.Name, "Expected name to be 'strings'")

		importPkg2, exists := pkg.Imports["github.com/example/stringss"]
		assert.True(t, exists, "Expected import to exist")

		assert.Equal(t, "stringss", importPkg2.Name, "Expected name to be 'stringss'")
	})

	t.Run("simple_collision", func(t *testing.T) {
		model := NewModel(nil)

		// Add first struct with import
		struct1 := &StructModel{
			Name: "User",
			Getters: []*GetterOutput{
				{
					Name: "GetName",
					Returns: []*ReturnOutput{
						{
							Value: &ValueOutput{
								TypePackage: &TypePackageOutput{
									Path:  "github.com/example/strings",
									Name:  "strings",
									Alias: "",
								},
							},
						},
					},
				},
			},
		}
		model.AddStruct("github.com/test/package1", "package1", struct1)

		// Add second struct with conflicting import name
		struct2 := &StructModel{
			Name: "Product",
			Getters: []*GetterOutput{
				{
					Name: "GetDescription",
					Returns: []*ReturnOutput{
						{
							Value: &ValueOutput{
								TypePackage: &TypePackageOutput{
									Path:  "github.com/other/strings",
									Name:  "strings", // Same name, different path
									Alias: "",
								},
							},
						},
					},
				},
			},
		}
		model.AddStruct("github.com/test/package1", "package1", struct2)

		// Add third struct with conflicting import name
		struct3 := &StructModel{
			Name: "Product",
			Getters: []*GetterOutput{
				{
					Name: "GetDescription",
					Returns: []*ReturnOutput{
						{
							Value: &ValueOutput{
								TypePackage: &TypePackageOutput{
									Path:  "github.com/another/strings",
									Name:  "strings", // Same name, different path
									Alias: "",
								},
							},
						},
					},
				},
			},
		}
		model.AddStruct("github.com/test/package1", "package1", struct3)

		// Get the package
		pkg := model.Packages["github.com/test/package1"]
		assert.NotNil(t, pkg, "Expected package to exist")

		// Check first import
		import1, exists := pkg.Imports["github.com/example/strings"]
		assert.True(t, exists, "Expected first import to exist")
		assert.Equal(t, "strings", import1.Name, "Expected first import alias to be 'strings'")

		// Check second import
		import2, exists := pkg.Imports["github.com/other/strings"]
		assert.True(t, exists, "Expected second import to exist")
		assert.Equal(t, "_strings", import2.Alias, "Expected second import alias to be '_strings'")

		import3, exists := pkg.Imports["github.com/another/strings"]
		assert.True(t, exists, "Expected third import to exist")
		assert.Equal(t, "__strings", import3.Alias, "Expected third import alias to be '_strings'")
	})
}
