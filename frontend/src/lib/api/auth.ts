import { cookies } from "next/headers";

const BASE_URL = process.env.API_URL;

export async function refreshToken(): Promise<string | null> {
	const cookieStore = await cookies();
	const refreshToken = cookieStore.get('refresh_token')?.value;

	if (!refreshToken) {
		return null;
	};

	try {
		const response = await fetch(`${BASE_URL}/auth/refresh`, {
			method: 'POST',
			headers: {
				Cookie: `refresh_token=${refreshToken}`,
			}
		});
		if (!response.ok) {
			return null;
		}

		const setCookieHeaders = response.headers.getSetCookie();

		if (!setCookieHeaders) {
			return null;
		};

		let newToken: null | string = null;

		setCookieHeaders.forEach((cookieString) => {
			const [nameValue] = cookieString.split(';');
			const [name, value] = nameValue.split('=');

			if (name === 'access_token') {
				newToken = value;
				cookieStore.set('access_token', value, {
					httpOnly: true,
					secure: process.env.NODE_ENV === 'production',
					path: '/',
					maxAge: 15 * 60,
				});
			}

			if (name === 'refresh_token') {
				cookieStore.set('refresh_token', value, {
					httpOnly: true,
					secure: process.env.NODE_ENV === 'production',
					path: '/',
					maxAge: 45 * 24 * 60 * 60,
				});
			}
		});

		return newToken as null | string;
	} catch (error) {
		return null;
	}
}
