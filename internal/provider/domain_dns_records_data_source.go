package provider

import (
	"context"
	"fmt"

	"github.com/MrLemur/migadu-go"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &DomainDNSRecordsDataSource{}

func NewDomainDNSRecordsDataSource() datasource.DataSource {
	return &DomainDNSRecordsDataSource{}
}

type DomainDNSRecordsDataSource struct {
	client *migadu.Client
}

type DomainDNSRecordsDataSourceModel struct {
	DomainName types.String `tfsdk:"domain_name"`
	Records    types.List   `tfsdk:"records"`
}

type DNSRecordModel struct {
	Type     types.String `tfsdk:"type"`
	Name     types.String `tfsdk:"name"`
	Value    types.String `tfsdk:"value"`
	Priority types.Int64  `tfsdk:"priority"`
	TTL      types.Int64  `tfsdk:"ttl"`
}

func (d *DomainDNSRecordsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_domain_dns_records"
}

func (d *DomainDNSRecordsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetches the required DNS records for a Migadu domain.",

		Attributes: map[string]schema.Attribute{
			"domain_name": schema.StringAttribute{
				MarkdownDescription: "The domain name.",
				Required:            true,
			},
			"records": schema.ListNestedAttribute{
				MarkdownDescription: "List of DNS records needed for the domain.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"type": schema.StringAttribute{
							MarkdownDescription: "DNS record type (e.g., MX, TXT, CNAME).",
							Computed:            true,
						},
						"name": schema.StringAttribute{
							MarkdownDescription: "DNS record name.",
							Computed:            true,
						},
						"value": schema.StringAttribute{
							MarkdownDescription: "DNS record value.",
							Computed:            true,
						},
						"priority": schema.Int64Attribute{
							MarkdownDescription: "DNS record priority (for MX records).",
							Computed:            true,
						},
						"ttl": schema.Int64Attribute{
							MarkdownDescription: "DNS record TTL.",
							Computed:            true,
						},
					},
				},
			},
		},
	}
}

func (d *DomainDNSRecordsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*migadu.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *migadu.Client, got: %T.", req.ProviderData),
		)
		return
	}

	d.client = client
}

func (d *DomainDNSRecordsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data DomainDNSRecordsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	domain := &migadu.Domain{Name: data.DomainName.ValueString()}
	records, err := d.client.GetDomainRecords(ctx, domain)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to get domain DNS records, got error: %s", err))
		return
	}

	var recordModels []DNSRecordModel
	for _, record := range records {
		recordModels = append(recordModels, DNSRecordModel{
			Type:     types.StringValue(record.Type),
			Name:     types.StringValue(record.Name),
			Value:    types.StringValue(record.Value),
			Priority: types.Int64Value(int64(record.Priority)),
			TTL:      types.Int64Value(int64(record.TTL)),
		})
	}

	recordsList, diags := types.ListValueFrom(ctx, types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"type":     types.StringType,
			"name":     types.StringType,
			"value":    types.StringType,
			"priority": types.Int64Type,
			"ttl":      types.Int64Type,
		},
	}, recordModels)
	resp.Diagnostics.Append(diags...)
	data.Records = recordsList

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
