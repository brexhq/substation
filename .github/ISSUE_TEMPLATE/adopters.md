name: Become an Adopter
about: Add the name of your organization to the list of adopters.
title: ''
labels: ''
assignees: ''

body:
  - type: markdown
    attributes:
      value: |
        Thank you for supporting Substation! By adding your organization to the list of adopters, you help raise awareness for the project and grow our community of users. Please fill out the information below to be added to the [list of adopters](https://github.com/brexhq/substation/blob/main/ADOPTERS.md).

  - type: input
    id: org-name
    attributes:
      label: Organization Name
      description: Name of your organization.
      placeholder: ex. Acme Corp
    validations:
      required: true
  - type: input
    id: org-url
    attributes:
      label: Organization Website
      description: Link to your organization's website.
      placeholder: ex. https://www.example.com
    validations:
      required: true
  - type: dropdown
      id: stage
      attributes:
        label: Stage of Adoption
        description: What is your current stage of adoption?
        options:
          - We're learning about Substation
          - We're testing Substation
          - We're using Substation in production
          - We're driving broad adoption of Substation
        default: 0
      validations:
        required: true
  - type: textarea
    id: use-case
    attributes:
      label: Description of Use
      description: Write 1 to 2 sentences about how your organization is using Substation.
    validations:
      required: true
