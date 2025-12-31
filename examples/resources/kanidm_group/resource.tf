# Example: Group with person and service account members
resource "kanidm_group" "developers" {
  id          = "developers"
  description = "Development team members"

  members = [
    kanidm_person.alice.id,
    kanidm_person.bob.id,
    kanidm_service_account.ci.id,
  ]
}

# Example: Group with only person members
resource "kanidm_group" "admins" {
  id          = "infrastructure-admins"
  description = "Infrastructure administrators with full access"

  members = [
    kanidm_person.alice_password.id,
    kanidm_person.charlie.id,
  ]
}

# Example: Empty group (members added later)
resource "kanidm_group" "monitoring" {
  id          = "monitoring-users"
  description = "Users with monitoring access"
}

# Example: Group with dynamic membership
resource "kanidm_group" "all_staff" {
  id          = "all-staff"
  description = "All staff members"

  members = [
    for person in kanidm_person.staff : person.id
  ]
}

# Example: Imported existing group
# Import command: terraform import kanidm_group.existing_group group_name
resource "kanidm_group" "existing_group" {
  id          = "existing"
  description = "Existing Group"
}
