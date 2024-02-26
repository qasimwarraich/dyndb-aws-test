resource "aws_dynamodb_table" "Countries" {
  name           = "Countries"
  billing_mode   = "PROVISIONED"
  read_capacity  = 25
  write_capacity = 25
  hash_key       = "Name"

  point_in_time_recovery {
    enabled = false
  }

  attribute {
    name = "Name"
    type = "S"
  }
}
