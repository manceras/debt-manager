import { NextRequest, NextResponse } from "next/server";
import { refreshTokensCore } from "./lib/api/auth-core";

const publicRoutes = ["/login", "/signup"];
const protectedRoutes = ["/app"];

export async function proxy(request: NextRequest) {
	const { pathname } = request.nextUrl;

	const isPublicRoute = publicRoutes.some((route) => pathname.startsWith(route));
	const isProtectedRoute = protectedRoutes.some((route) => pathname.startsWith(route));

	const accessToken = request.cookies.get("access_token")?.value;
	const refreshToken = request.cookies.get("refresh_token")?.value;

	if (isPublicRoute && accessToken) {
		return NextResponse.redirect(new URL("/app", request.url));
	}

	if (isProtectedRoute) {
		if (!refreshToken && !accessToken) {
			return NextResponse.redirect(new URL("/login", request.url));
		}

		if (!accessToken && refreshToken) {
			const tokens = await refreshTokensCore(refreshToken);
			if (!tokens) {
				return NextResponse.redirect(new URL("/login", request.url));
			}

			const response = NextResponse.next();
			response.cookies.set("access_token", tokens.accessToken, {
				httpOnly: true,
				secure: process.env.NODE_ENV === "production",
				path: "/",
				maxAge: 15 * 60,
			});

			if (tokens.refreshToken) {
				response.cookies.set("refresh_token", tokens.refreshToken, {
					httpOnly: true,
					secure: process.env.NODE_ENV === "production",
					path: "/auth/refresh",
					maxAge: 45 * 24 * 60 * 60,
				});
			}

			return response;
		}
	}

	return NextResponse.next();
}

export const config = {
	matcher: ['/((?!api|_next/static|_next/image|favicon.ico).*)'],
};
