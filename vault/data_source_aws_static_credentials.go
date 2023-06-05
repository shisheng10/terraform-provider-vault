package vault

import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-vault/internal/consts"
	"github.com/hashicorp/terraform-provider-vault/internal/provider"
)

const staticCredsAffix = "static-creds"

func awsStaticCredDataSource() *schema.Resource {
	return &schema.Resource{
		ReadContext: ReadContextWrapper(awsStaticCredentialsDataSourceRead),
		Schema: map[string]*schema.Schema{
			consts.FieldMount: {
				Type:        schema.TypeString,
				Required:    true,
				Description: "AWS Secret Backend to read credentials from.",
			},
			consts.FieldName: {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Name of the role.",
			},
			consts.FieldAWSAccessKeyID: {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "AWS access key ID read from Vault.",
				Sensitive:   true,
			},
			consts.FieldAWSSecretAccessKey: {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "AWS secret key read from Vault.",
				Sensitive:   true,
			},
		},
	}
}

func awsStaticCredentialsDataSourceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, err := provider.GetClient(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	mount := d.Get(consts.FieldMount).(string)
	role := d.Get(consts.FieldName).(string)
	fullPath := fmt.Sprintf("%s/%s/%s", mount, staticCredsAffix, role)

	secret, err := client.Logical().ReadWithContext(ctx, fullPath)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error reading from Vault: %s", err))
	}
	log.Printf("[DEBUG] Read %q from Vault", fullPath)
	if secret == nil {
		return diag.FromErr(fmt.Errorf("no role found at %q", fullPath))
	}

	d.SetId(fullPath)
	if err := d.Set(consts.FieldAWSAccessKeyID, secret.Data["access_key_id"]); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set(consts.FieldAWSSecretAccessKey, secret.Data["secret_access_key"]); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
