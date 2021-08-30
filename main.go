package main

import (
	"blackdark-aws-provider/blackdarkaws"
	"context"

	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

func main() {
	tfsdk.Serve(context.Background(), blackdarkaws.New, tfsdk.ServeOpts{
		Name: "blackdark-aws",
	})
}
