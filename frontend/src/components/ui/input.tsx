import * as React from "react"

import { cn } from "@/lib/utils"
import { Eye, EyeClosed } from "lucide-react";
import { Button } from "./button";

const commonClassName: string = cn(
	"file:text-foreground placeholder:text-muted-foreground selection:bg-primary selection:text-primary-foreground dark:bg-input/30 border-input h-9 w-full min-w-0 rounded-md border bg-transparent px-3 py-1 text-base shadow-xs transition-[color,box-shadow] outline-none file:inline-flex file:h-7 file:border-0 file:bg-transparent file:text-sm file:font-medium disabled:pointer-events-none disabled:cursor-not-allowed disabled:opacity-50 md:text-sm",
	"focus-visible:border-ring focus-visible:ring-ring/50 focus-visible:ring-[3px]",
	"aria-invalid:ring-destructive/20 dark:aria-invalid:ring-destructive/40 aria-invalid:border-destructive",
)

function Input({ className, type, ...props }: React.ComponentProps<"input">) {
	return (
		<input
			type={type}
			data-slot="input"
			className={cn(
				commonClassName,
				className
			)}
			{...props}
		/>
	)
}

function PasswordInputWithReveal(props: React.ComponentProps<"input">) {
	const [visible, setVisible] = React.useState(false);

	return (
		<div className="relative w-full">
			<input
				type={visible ? "text" : "password"}
				data-slot="input"
				className={cn(
					commonClassName,
					"pr-10",
					props.className
				)}
				{...props}
			/>
			<Button
				type="button"
				onClick={() => setVisible(!visible)}
				className="absolute right-2 top-1/2 -translate-y-1/2 h-7 w-7"
				variant="ghost"
				tabIndex={-1}
			>
				{visible ? <EyeClosed /> : <Eye />}
			</Button>
		</div>
	);
}

export { Input, PasswordInputWithReveal }
