import { CircleCheck, Info, TriangleAlert } from "lucide-react";
import { Card, CardContent } from "./card";
import { cn } from "@/lib/utils";
import { ClassValue } from "clsx";

type Variant = "success" | "error" | "warning" | "info";

type AlertCardProps = {
	variant?: Variant;
	title: string;
	description?: string;
};

const styles: Record<Variant, {
	icon: React.ComponentType<any>;
	styles: ClassValue;
}> = {
	success: {
		icon: CircleCheck,
		styles: "border-green-400 text-green-800",
	},
	error: {
		icon: TriangleAlert,
		styles: "border-red-400 text-red-800",
	},
	warning: {
		icon: TriangleAlert,
		styles: "border-yellow-400 text-yellow-800",
	},
	info: {
		icon: Info,
		styles: "border-neutral-400 text-neutral-800",
	}
}

export default function AlertCard({ variant = "info", title, description, className, ...props }: AlertCardProps & React.ComponentProps<"div">) {
	const { icon: Icon, styles: variantStyles } = styles[variant];

	return (
		<Card
			className={cn(variantStyles, className)}
			{...props}
		>
			<CardContent className="flex flex-row gap-3">
				<Icon className="h-5 w-5 shrink-0 mt-1" />
				<div className="flex-1 flex-col gap-2">
					<p className="font-medium">{title}</p>
					{description && <p className="text-sm text-neutral-600">{description}</p>}
				</div>
			</CardContent>
		</Card>
	);
}
