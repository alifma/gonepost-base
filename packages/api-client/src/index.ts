import createClient from "openapi-fetch";
import type { paths } from "./types.gen";

export type { paths, components } from "./types.gen";

/**
 * Typed HTTP client for the Gonepost API, generated from
 * apps/api/openapi/openapi.yaml. This is the only allowed way for
 * apps/web to talk to apps/api — never fetch() the API directly.
 */
export function createApiClient(baseUrl: string) {
  return createClient<paths>({ baseUrl });
}
