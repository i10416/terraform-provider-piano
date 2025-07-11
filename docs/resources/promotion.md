---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "piano_promotion Resource - piano"
subcategory: ""
description: |-
  Promotion represents a special discount. Users can use a promotion code associated with a promotion to get a discount.For more details, see https://docs.piano.io/promotions/
---

# piano_promotion (Resource)

Promotion represents a special discount. Users can use a promotion code associated with a promotion to get a discount.For more details, see https://docs.piano.io/promotions/

## Example Usage

```terraform
resource "piano_promotion" "sample" {
  aid  = "sample-aid"
  name = "sample"
  # null indicates unlimited uses
  uses_allowed = null
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `aid` (String) The application ID
- `name` (String) The promotion name
- `term_dependency_type` (String) The type of dependency to terms.
When the value is "all", the promotion can be applied to app terms.
When the value is "include", the promotion can be applied to those specific terms.
When the value is "unlocked", the promotion allows customers to access special terms that they could not have accessed without the code

### Optional

- `apply_to_all_billing_periods` (Boolean) Whether to apply the promotion discount to all billing periods ("TRUE")or the first billing period only ("FALSE")
- `billing_period_limit` (Number) Promotion discount applies to number of billing periods
- `can_be_applied_on_renewal` (Boolean) Whether the promotion can be applied on renewal
- `discount_type` (String) The promotion discount type
- `end_date` (Number) The end date
- `fixed_promotion_code` (String) The fixed value for all the promotion codes
- `never_allow_zero` (Boolean) Never allow the value of checkout to be zero
- `new_customers_only` (Boolean) Whether the promotion allows new customers only
- `percentage_discount` (Number) The promotion discount, percentage
- `promotion_code_prefix` (String) The prefix for all the codes
- `start_date` (Number) The start date.
- `uses_allowed` (Number) The number of uses allowed by the promotion. If this value is null, it indicates unlimited uses allowed.

### Read-Only

- `create_date` (Number) The creation date
- `fixed_discount_list` (Attributes List) (see [below for nested schema](#nestedatt--fixed_discount_list))
- `promotion_id` (String) The promotion ID
- `unlimited_uses` (Boolean) Whether to allow unlimited uses
- `update_date` (Number) The update date

<a id="nestedatt--fixed_discount_list"></a>
### Nested Schema for `fixed_discount_list`

Read-Only:

- `amount` (String) The fixed discount amount
- `amount_value` (Number) The fixed discount amount value
- `currency` (String) The currency of the fixed discount
- `fixed_discount_id` (String) The fixed discount ID

## Import

Import is supported using the following syntax:

```shell
terraform import piano_promotion.sample aid/promotion_id
```
