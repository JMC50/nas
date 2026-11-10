export interface OidcAddresses {
  loginUrl: string;
  logoutUrl: string;
}

// Null when the server uses another sign-in mode, or when it does not answer.
export async function oidcAddresses(): Promise<OidcAddresses | null> {
  try {
    const response = await fetch("/server/auth/config");
    const config = await response.json();
    if (config.authType !== "oidc") return null;
    return { loginUrl: config.oidcLoginUrl, logoutUrl: config.oidcLogoutUrl };
  } catch {
    return null;
  }
}

export async function signOutAddress(fallback: string): Promise<string> {
  const oidc = await oidcAddresses();
  return oidc ? oidc.logoutUrl : fallback;
}
