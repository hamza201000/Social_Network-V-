import Bar from "@/components/bar";


export default function RootLayoutDashboardLayout({
    children,
}: Readonly<{
    children: React.ReactNode;
}>) {
    return (
        <html
            lang="en"
        >
            <body >
            <Bar/>
            {children}</body>
        </html>
    );
}