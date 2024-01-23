import { useMutation } from "react-query";
import {
  CompileJSRequest,
  CompileJSResponse,
  DeploymentCreateRequest,
  DeploymentCreateResponse,
} from "./wire";
import { APIResponse } from "./base";

function handleApiResponse<T extends APIResponse<any>>(
  resp: Promise<T>
): Promise<T> {
  return resp;
}

export function useCompileJsMutation() {
  return useMutation((req: CompileJSRequest) => {
    return fetch(`/api/v1/compile/js`, {
      method: "POST",
      body: JSON.stringify(req),
      headers: {
        "Content-Type": "application/json",
      },
    }).then((res) => handleApiResponse<CompileJSResponse>(res.json()));
  });
}

export function useDeploymentCreateMutation() {
  return useMutation((req: DeploymentCreateRequest) => {
    return fetch(`/api/v1/guilds/1234/deployments`, {
      method: "POST",
      body: JSON.stringify(req),
      headers: {
        "Content-Type": "application/json",
      },
    }).then((res) => handleApiResponse<DeploymentCreateResponse>(res.json()));
  });
}
