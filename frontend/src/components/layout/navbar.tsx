"use client";

import { User } from "@/types/user";
import { ChevronLeft, UserIcon } from "lucide-react";

interface NavbarProps {
	user: User;
}

export default function Navbar({ user }: NavbarProps) {
	return (
		<nav className="w-full h-16 flex items-center justify-between px-4">
			<ChevronLeft />
			<h1>Title</h1>
			<div><UserIcon />{user.username}</div>
		</nav>
	);
}
