import Bar from "@/components/layout/bar";


export default function DashboardLayout({
    children,
}: Readonly<{
    children: React.ReactNode;
}>) {
    return (
        <>
            <Bar />
            {children}
        </>
    );
}