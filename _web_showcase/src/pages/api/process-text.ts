import type { APIRoute } from "astro";
import { runJSoon } from "../../lib/jsoon";

export const POST: APIRoute = async ({ request }) => {
  try {
    const { text, options } = await request.json();
    const data = runJSoon(text, options);
    return new Response(JSON.stringify(data), {
      status: 200,
      headers: { "Content-Type": "application/json" },
    });
  } catch (error: any) {
    return new Response(JSON.stringify({ error: error.message }), {
      status: 500,
      headers: { "Content-Type": "application/json" },
    });
  }
};
