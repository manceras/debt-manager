import { cookies } from "next/headers";
import { redirect } from "next/navigation";
import { refreshTokensCore } from "./auth-core";
const BASE_URL = process.env.API_URL;

export async function apiClient<TResponse, TBody = undefined>(
	path: string,
	options: RequestInit & { body?: TBody | FormData } = {}
): Promise<TResponse> {

	const cookieStore = await cookies();
	const accessToken = cookieStore.get('access_token')?.value;
	const refreshToken = cookieStore.get('refresh_token')?.value;

	const getHeaders = (tempAccess?: string, tempRefresh?: string) => {
		const headers: Record<string, string> = {
			...((options.headers as Record<string, string>) || {}),
		};

		const currentAccess = tempAccess || accessToken;
		const currentRefresh = tempRefresh || refreshToken;

		const cookieParts: string[] = [];
		if (currentAccess) cookieParts.push(`access_token=${currentAccess}`);
		if (currentRefresh) cookieParts.push(`refresh_token=${currentRefresh}`);

		if (cookieParts.length > 0) {
			headers['Cookie'] = cookieParts.join('; ');
		}

		if (!(options.body instanceof FormData)) {
			headers['Content-Type'] = 'application/json';
		}

		return headers;
	};

	const getBody = (body?: TBody | FormData) => {
		if (!body) return undefined;
		if (body instanceof FormData) return body;
		return JSON.stringify(body);
	};

	let res = await fetch(`${BASE_URL}${path}`, {
		...options,
		headers: getHeaders(),
		body: getBody(options.body),
	});

	if (res.status === 401) {
		if (!refreshToken) {
			redirect('/login');
		}

		const newTokens = await refreshTokensCore(refreshToken);

		if (!newTokens) {
			redirect('/login');
		}


		try {
			cookieStore.set('access_token', newTokens.accessToken, {
				httpOnly: true, secure: process.env.NODE_ENV === 'production', path: '/', maxAge: 15 * 60
			});
			if (newTokens.refreshToken) {
				cookieStore.set('refresh_token', newTokens.refreshToken, {
					httpOnly: true, secure: process.env.NODE_ENV === 'production', path: '/', maxAge: 45 * 24 * 60 * 60
				});
			}
		} catch (e) {
		}

		res = await fetch(`${BASE_URL}${path}`, {
			...options,
			headers: getHeaders(newTokens.accessToken, newTokens.refreshToken),
			body: getBody(options.body),
		});
	}

	if (!res.ok) {
		const errorData = await res.json().catch(() => ({}));
		throw new Error(errorData.message || `API Error: ${res.status}`);
	}

	return res.json() as Promise<TResponse>;
}
