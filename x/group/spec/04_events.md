# Events

The group module emits the following events:

## EventCreateGroup

| Type                                  | Attribute Key | Attribute Value                       |
|---------------------------------------|---------------|---------------------------------------|
| message                               | action        | /fetchai.group.v1alpha1.Msg/CreateGroup |
| fetchai.group.v1alpha1.EventCreateGroup | group_id      | {groupId}                             |

## EventUpdateGroup

| Type                                  | Attribute Key | Attribute Value                                                 |
|---------------------------------------|---------------|-----------------------------------------------------------------|
| message                               | action        | /fetchai.group.v1alpha1.Msg/UpdateGroup{Admin\|Metadata\|Members} |
| fetchai.group.v1alpha1.EventUpdateGroup | group_id      | {groupId}                                                       |

## EventCreateGroupAccount

| Type                                         | Attribute Key | Attribute Value                              |
|----------------------------------------------|---------------|----------------------------------------------|
| message                                      | action        | /fetchai.group.v1alpha1.Msg/CreateGroupAccount |
| fetchai.group.v1alpha1.EventCreateGroupAccount | address       | {groupAccountAddress}                        |

## EventUpdateGroupAccount

| Type                                         | Attribute Key | Attribute Value                                                               |
|----------------------------------------------|---------------|-------------------------------------------------------------------------------|
| message                                      | action        | /fetchai.group.v1alpha1.Msg/UpdateGroupAccount{Admin\|Metadata\|DecisionPolicy} |
| fetchai.group.v1alpha1.EventUpdateGroupAccount | address       | {groupAccountAddress}                                                         |

## EventCreateProposal

| Type                                     | Attribute Key | Attribute Value                          |
|------------------------------------------|---------------|------------------------------------------|
| message                                  | action        | /fetchai.group.v1alpha1.Msg/CreateProposal |
| fetchai.group.v1alpha1.EventCreateProposal | proposal_id   | {proposalId}                             |

## EventVote

| Type                           | Attribute Key | Attribute Value                |
|--------------------------------|---------------|--------------------------------|
| message                        | action        | /fetchai.group.v1alpha1.Msg/Vote |
| fetchai.group.v1alpha1.EventVote | proposal_id   | {proposalId}                   |

## EventExec

| Type                           | Attribute Key | Attribute Value                |
|--------------------------------|---------------|--------------------------------|
| message                        | action        | /fetchai.group.v1alpha1.Msg/Exec |
| fetchai.group.v1alpha1.EventExec | proposal_id   | {proposalId}                   |