import { cookies } from "next/headers";
import { refreshToken } from "./auth";
import { redirect } from "next/navigation";

const BASE_URL = process.env.API_URL;

export async function apiClient<R, B = undefined>(
	url: string,
	options: RequestInit & { body?: B } = {},
): Promise<R> {
	const cookieStore = await cookies();
	const token = cookieStore.get('access_token')?.value;

	const getHeaders = (t?: string) => {
		const headers: Record<string, string> = {
			...(t && { Authorization: `Bearer ${t}` }),
			...((options.headers as Record<string, string>) || {}),
		};

		if (!(options.body instanceof FormData)) {
			headers['Content-Type'] = 'application/json';
		}

		return headers;
	};

	const getBody = (body?: B | FormData) => {
		if (!body) return undefined;
		if (body instanceof FormData) return body;
		return JSON.stringify(body);
	};

	let res = await fetch(`${BASE_URL}${url}`, {
		...options,
		headers: getHeaders(token),
		body: getBody(options.body),
	});

	if (res.status === 401 && token) {
		const newToken = await refreshToken();
		if (!newToken) {
			redirect('/login');
		}

		res = await fetch(`${BASE_URL}${url}`, {
			...options,
			headers: getHeaders(newToken),
			body: getBody(options.body),
		});

		if (res.status === 401) {
			redirect('/login');
		}

		if (!res.ok) {
			const errorData = await res.json();
			throw new Error(errorData.message || `API request failed after token refresh, status: ${res.status}`);
		}
	}

	if (!res.ok) {
		const errorData = await res.json();
		throw new Error(errorData.message || `API request failed, status: ${res.status}`);
	}

	return res.json();
}


