"use client";

import { User } from "@/types/user";
import { ChevronLeft, LogOut, Moon, Sun, SunMoon, UserIcon } from "lucide-react";
import { DropdownMenu, DropdownMenuContent, DropdownMenuGroup, DropdownMenuItem, DropdownMenuLabel, DropdownMenuSeparator, DropdownMenuTrigger } from "../ui/dropdown-menu";
import { Button } from "../ui/button";
import { logout } from "@/lib/api/auth";
import { useTheme } from "next-themes";
import { useTitle } from "@/context/title-context";
import { usePathname } from "next/navigation";
import Link from "next/link";

const BACK_MAP: Record<string, string | null> = {
	"/": null,
	"/list": "/"
};

function getBackPath(pathname: string): string | null {
	return Object.keys(BACK_MAP).find((path) => pathname.startsWith("/app" + path)) ?? null;
}

interface NavbarProps {
	user: User;
}

export default function Navbar({ user }: NavbarProps) {
	const { setTheme, theme } = useTheme();
	const { title } = useTitle();
	const pathname = usePathname();

	const backPath = getBackPath(pathname);

	return (
		<nav className="w-full h-16 flex items-center justify-between px-4">
			<div className="h-9 w-9 flex items-center justify-center">
				{ backPath && (<Button variant="ghost" size="icon" asChild>
					<Link href={backPath}>
						<ChevronLeft />
					</Link>
				</Button>) }
			</div>
			<h1>{title}</h1>
			<DropdownMenu>
				<DropdownMenuTrigger asChild>
					<button className="rounded-full bg-secondary p-1 cursor-pointer h-9 w-9 flex items-center justify-between"><UserIcon /></button>
				</DropdownMenuTrigger>
				<DropdownMenuContent>
					<DropdownMenuLabel>{user.username}</DropdownMenuLabel>
					<DropdownMenuLabel className="text-sm text-muted-foreground flex-row">{user.email}</DropdownMenuLabel>
					<DropdownMenuSeparator />
					<DropdownMenuGroup>
						<DropdownMenuItem onSelect={() => logout()}>
							<LogOut />
							Logout
						</DropdownMenuItem>
					</DropdownMenuGroup>
					<DropdownMenuSeparator />
					<DropdownMenuGroup>
						<DropdownMenu>
							<DropdownMenuTrigger asChild>
								<Button variant="outline" size="sm">
									{theme === "light" && (<>
										<Sun className="mr-2" />
										<p>Light</p>
									</>)}
									{theme === "dark" && (<>
										<Moon className="mr-2" />
										<p>Dark</p>
									</>
									)}
									{theme === "system" && (<>
										<SunMoon className="mr-2" />
										<p>System</p>
									</>)}
								</Button>
							</DropdownMenuTrigger>
							<DropdownMenuContent align="end">
								<DropdownMenuItem onClick={() => setTheme("light")}>
									Light
								</DropdownMenuItem>
								<DropdownMenuItem onClick={() => setTheme("dark")}>
									Dark
								</DropdownMenuItem>
								<DropdownMenuItem onClick={() => setTheme("system")}>
									System
								</DropdownMenuItem>
							</DropdownMenuContent>
						</DropdownMenu>
					</DropdownMenuGroup>
				</DropdownMenuContent>
			</DropdownMenu>
		</nav>
	);
}
