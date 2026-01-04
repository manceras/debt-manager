'use server';

import { cookies } from 'next/headers';
import { redirect } from 'next/navigation';

const BASE_URL = process.env.API_URL;

console.log('API_URL:', BASE_URL);

export type LoginState = {
	error?: string;
	message?: string;
}

export async function loginAction(_: LoginState, formData: FormData): Promise<LoginState> {
	const email = formData.get('email') as string;
	const password = formData.get('password') as string;

	if (!email || !password) {
		return { error: 'Email and password are required.' };
	}

	try {
		const response = await fetch(`${BASE_URL}/auth/login`, {
			method: 'POST',
			headers: {
				'Content-Type': 'application/json',
			},
			body: JSON.stringify({ email, password }),
		});

		if (!response.ok) {
			const errorData = await response.json();
			if (response.status === 401) {
				return { error: 'Invalid email or password.' };
			}
			return { error: errorData.error || 'Login failed.' };
		}

		const setCookieHeader = response.headers.getSetCookie();

		if (!setCookieHeader || setCookieHeader.length === 0) {
			return { error: 'Login successfull but no session created.' };
		}

		const cookieStore = await cookies();

		setCookieHeader.forEach((cookieString) => {
			const [cookiePair] = cookieString.split(';');
			const [name, value] = cookiePair.split('=');

			if (name === "access_token" || name === "refresh_token") {
				cookieStore.set(name, value, {
					httpOnly: true,
					secure: process.env.NODE_ENV === 'production',
					path: name === "access_token" ? "/" : "/",
					maxAge: name === "access_token" ? 15 * 60 : 45 * 24 * 60 * 60,
				});
			}
		});
	} catch (error) {
		console.error('Login error:', error);
		return { error: 'An unexpected error occurred. Please try again.' };
	}

	redirect('/app');
}
