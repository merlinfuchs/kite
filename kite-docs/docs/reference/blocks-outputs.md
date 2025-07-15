---
sidebar_position: 7
---

# Block Outputs

Many blocks in Kite return values that can be reused later in the flow using the format `{{nodes.nodeName.result}}`.

The table below describes some of the most common block outputs and how to use them:

| Block                                           | Output                                                                 |
|------------------------------------------------|-------------------------------------------------------------------------|
| Create response message / Create channel message | Add `.id` to the placeholder to get the created message ID            |
| Set Variable / Get Variable                    | Returns the value of the selected variable                              |
