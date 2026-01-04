"use server";

import { cookies } from "next/headers";
import { refreshTokensCore } from "./auth-core";
import { redirect } from "next/navigation";

const BASE_URL = process.env.API_URL;

export async function refreshToken(): Promise<string | null> {
    const cookieStore = await cookies();
    const oldRefreshToken = cookieStore.get('refresh_token')?.value;

    if (!oldRefreshToken) return null;

    // Use the core logic
    const tokens = await refreshTokensCore(oldRefreshToken);

    if (!tokens) return null;

    // Apply side effects (Setting cookies in Node.js context)
    cookieStore.set('access_token', tokens.accessToken, {
        httpOnly: true,
        secure: process.env.NODE_ENV === 'production',
        path: '/',
        maxAge: 15 * 60,
    });

    if (tokens.refreshToken) {
        cookieStore.set('refresh_token', tokens.refreshToken, {
            httpOnly: true,
            secure: process.env.NODE_ENV === 'production',
            path: '/',
            maxAge: 45 * 24 * 60 * 60,
        });
    }

    return tokens.accessToken;
}

export async function logout() {
	const cookieStore = await cookies();
	try {
		const refreshToken = cookieStore.get("refresh_token")?.value;
		let accessToken = cookieStore.get("access_token")?.value;

		if (refreshToken && !accessToken) {
			const response = await fetch(`${BASE_URL}/auth/refresh`, {
				method: "POST",
				headers: {
					Cookie: `refresh_token=${refreshToken}`,
				}
			});

			if (response.ok) {
				const cookiesStore = await cookies();
				accessToken = cookiesStore.get("access_token")?.value;
			}
		}

		if (accessToken) {
			await fetch(`${BASE_URL}/auth/logout`, {
				method: "POST",
				headers: {
					Cookie: `access_token=${accessToken}; refresh_token=${refreshToken}`,
				}
			});
		}
	} finally {
		cookieStore.delete("access_token");
		cookieStore.delete("refresh_token");

		redirect("/login");
	}
}
