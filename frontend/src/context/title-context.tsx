"use client";

import { createContext, useContext, useState } from "react";


interface TitleValue {
	title: string;
	setTitle: (title: string) => void;
}

const TitleContext = createContext<TitleValue | null>(null);

export function useTitle(): TitleValue {
	const context = useContext(TitleContext);
	if (!context) {
		throw new Error("useTitleContext must be used within a TitleProvider");
	}
	return context;
}

export function TitleProvider({ children }: { children: React.ReactNode }) {
	const defaultTitle = "Debt manager";
	const [title, setStateTitle] = useState(defaultTitle);

	function setTitle(newTitle: string) {
		setStateTitle(newTitle || defaultTitle);
	}

	return (
		<TitleContext.Provider value={{ title, setTitle }}>
			{children}
		</TitleContext.Provider>
	);
}
