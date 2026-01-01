import { cookies } from "next/headers";
import { refreshTokensCore } from "./auth-core";

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

