type ApiError = {
  error?: { code?: string; message?: string; fields?: { field: string; message: string }[] };
};

/** Human-readable message out of an API error body. */
export function errorMessage(err: unknown, fallback = 'Something went wrong') {
  const e = (err as ApiError | undefined)?.error;
  if (e?.fields?.length) return e.fields.map((f) => f.message).join(', ');
  return e?.message ?? fallback;
}

/**
 * Turns an openapi-fetch result into "data or throw", so react-query sees
 * failures as errors. The thrown error carries the API message.
 */
export async function unwrap<T>(
  request: Promise<{ data?: T; error?: unknown; response: Response }>
): Promise<T> {
  const { data, error, response } = await request;
  if (error !== undefined) throw new Error(errorMessage(error));
  if (data === undefined) {
    if (response.ok) return undefined as T; // 204 No Content
    throw new Error(errorMessage(undefined));
  }
  return data;
}
