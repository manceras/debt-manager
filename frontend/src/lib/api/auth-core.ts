const BASE_URL = process.env.API_URL;

export interface TokenResponse {
  accessToken: string;
  refreshToken?: string;
}

export async function refreshTokensCore(refreshToken: string): Promise<TokenResponse | null> {
  try {
    const response = await fetch(`${BASE_URL}/auth/refresh`, {
      method: 'POST',
      headers: {
        Cookie: `refresh_token=${refreshToken}`,
      },
    });

    if (!response.ok) return null;

    const setCookieHeaders = response.headers.getSetCookie();
    if (!setCookieHeaders || setCookieHeaders.length === 0) return null;

    let accessToken = '';
    let newRefreshToken: string | undefined;

    setCookieHeaders.forEach((cookieString) => {
      const [nameValue] = cookieString.split(';');
      const [name, value] = nameValue.split('=');

      if (name === 'access_token') accessToken = value;
      if (name === 'refresh_token') newRefreshToken = value;
    });

    if (!accessToken) return null;

    return { accessToken, refreshToken: newRefreshToken };
  } catch (error) {
    console.error("Refresh core failed", error);
    return null;
  }
}
