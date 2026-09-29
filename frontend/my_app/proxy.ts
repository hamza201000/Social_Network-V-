

import { NextResponse } from "next/server";
import type { NextRequest } from "next/server";


export function proxy(request: NextRequest) {
    const session = request.cookies.get("session_id");
    console.log("sesion:::",session);

    const { pathname } = request.nextUrl
    console.log("pathname",pathname);
    
    if (session && (pathname === "/login" || pathname === "/register")) {
        return NextResponse.redirect(new URL("/", request.url));
    }
    if (!session && !(pathname === "/login" || pathname === "/register")) {
        return NextResponse.redirect(new URL("/login", request.url));
    }

    return NextResponse.next();
}

export const config = {
  matcher: [
    "//:path*",
    "/login",
    "/register",
  ],
};
