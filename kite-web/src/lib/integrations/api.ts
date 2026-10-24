// The endpoints of an integration's API that Kite knows about, in one format
// for every integration. It's generated from the integration's OpenAPI spec
// where there is one, like scripts/discord-api.mjs does for Discord, and
// written by hand otherwise. blocks.test.ts checks request blocks against it.

export interface ApiParam {
  name: string;
  // snowflake, integer, number, boolean, array or string
  type: string;
  required: boolean;
}

export interface ApiOperation {
  // operationId in the integration's spec.
  id: string;
  method: string;
  path: string;
  path_params: ApiParam[];
  query_params: ApiParam[];
  has_body: boolean;
  // Top-level properties of a JSON object body. Missing for bodies that are
  // lists or unions.
  body_params?: ApiParam[];
  // Discord only: the endpoint acts on the interaction the flow runs with, and
  // the service fills in {application_id} and {interaction_token} of its path.
  interaction?: boolean;
}

export interface IntegrationApi {
  // Where the operations come from: a spec at a pinned commit, or a note
  // that they're written by hand.
  source: string;
  operations: ApiOperation[];
}
