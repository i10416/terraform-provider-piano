---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "piano_promotion Data Source - piano"
subcategory: ""
description: |-
  
---

# piano_promotion (Data Source)





<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `aid` (String) The application ID
- `promotion_id` (String) The promotion ID

### Optional

- `fixed_promotion_code` (String) The fixed value for all the promotion codes
- `promotion_code_prefix` (String) The prefix for all the codes
- `uses_allowed` (Number) The number of uses allowed by the promotion

### Read-Only

- `apply_to_all_billing_periods` (Boolean) Whether to apply the promotion discount to all billing periods ("TRUE")or the first billing period only ("FALSE")
- `billing_period_limit` (Number) Promotion discount applies to number of billing periods
- `can_be_applied_on_renewal` (Boolean) Whether the promotion can be applied on renewal
- `create_by` (String) The user who created the object
- `create_date` (Number) The creation date
- `deleted` (Boolean) Whether the object is deleted
- `discount` (String) The promotion discount, formatted
- `discount_amount` (Number) The promotion discount
- `discount_currency` (String) The promotion discount currency
- `discount_type` (String) The promotion discount type
- `end_date` (Number) The end date
- `fixed_discount_list` (Attributes List) (see [below for nested schema](#nestedatt--fixed_discount_list))
- `name` (String) The promotion name
- `never_allow_zero` (Boolean) Never allow the value of checkout to be zero
- `new_customers_only` (Boolean) Whether the promotion allows new customers only
- `percentage_discount` (Number) The promotion discount, percentage
- `start_date` (Number) The start date.
- `status` (String) The promotion status
- `term_dependency_type` (String) The type of dependency to terms
- `unlimited_uses` (Boolean) Whether to allow unlimited uses
- `update_by` (String) The last user to update the object
- `update_date` (Number) The update date
- `uses` (Number) How many times the promotion has been used

<a id="nestedatt--fixed_discount_list"></a>
### Nested Schema for `fixed_discount_list`

Read-Only:

- `amount` (String) The fixed discount amount
- `amount_value` (Number) The fixed discount amount value
- `currency` (String) The currency of the fixed discount
- `fixed_discount_id` (String) The fixed discount ID
