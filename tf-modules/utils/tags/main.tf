locals {
  # Configure default tags and merge them with the input tags.
  default_tags = {
    # NOTE: GCP requires labels to start with a letter, all letters to be in
    # lowercase, doesn't allow colons(':') and spaces. Use a custom time
    # format to be compatible with GCP labels.
    # If needed in the future, timezone information can also be added.
    #
    # To parse this into a valid time format, remove the first filler letter
    # 'x' and the last letter 's', replace all the remaining letters with ':'
    # and underscore ('_') with space.
    # For example, convert "x2023-04-22_10h05m15s" into "2023-04-22 10:05:15"
    # and parse it with any time parser.
    #
    # A Go reference implementation of this parser is available in tftestenv
    # packages: ParseCreatedAtTime().
    createdat = formatdate("'x'YYYY-MM-DD_hh'h'mm'm'ss's'", timestamp())

    test = "true"
  }
  tags = merge(local.default_tags, var.tags)
}
