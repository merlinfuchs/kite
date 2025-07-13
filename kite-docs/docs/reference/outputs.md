---
sidebar_position: 7
---

# Block Outputs

Many blocks in Kite return values that can be reused later in the flow using the format `{{nodes.nodeName.result}}`.

The table below describes some of the most common block outputs and how to use them:

<table class="custom-table">
  <thead>
    <tr>
      <th>Block</th>
      <th>Output</th>
    </tr>
  </thead>
  <tbody>
    <tr>
      <td>Create response message / Create channel message</td>
      <td>Add <span class="inline-code-snippet">.id</span> to the placeholder to get the created message ID</td>
    </tr>
    <tr>
      <td>Set Variable / Get Variable</td>
      <td>Returns the value of the selected variable</td>
    </tr>
  </tbody>
</table>
