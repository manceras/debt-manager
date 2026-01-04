import Navbar from "@/components/layout/navbar";
import { TitleProvider } from "@/context/title-context";
import { apiClient } from "@/lib/api/client";
import { User } from "@/types/user";

export default async function AppLayout({ children }: { children: React.ReactNode }) {
	let user: User | null = null;
	try {
		user = await apiClient<User>("/auth/me");
	} catch (error) {
		console.error("Failed to fetch user:", error);
		user = null;
	}

	if (!user) {
		return <div>Please log in to access the application.</div>;
	}

	return (
		<TitleProvider>
			<div className="flex-col h-dvh w-dvw">
				<Navbar user={user} />
				<main className="flex-1 overflow-auto">{children}</main>
			</div>
		</TitleProvider>
	)
}
