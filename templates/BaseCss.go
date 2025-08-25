package templates

var BaseCss = `
		// Base CSS for all emails
		@import url('https://fonts.googleapis.com/css2?family=Inter:wght@300;400;500;600;700&display=swap');
		* {
			margin: 0;
			padding: 0;
			box-sizing: border-box;
		}
		body {
			font-family: 'Inter', -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
			line-height: 1.6;
			color: #0f172a; /* --card-foreground */
			background-color: #f1f5f9; /* Light background for contrast */
			padding: 20px 0;
		}
		.email-wrapper {
			max-width: 600px;
			margin: 0 auto;
			background: #ffffff;
			border-radius: 12px;
			overflow: hidden;
			box-shadow: 0 4px 12px rgba(0, 0, 0, 0.05);
		}
		.header {
			background-color: #00434c; /* --secondary */
			padding: 30px 25px;
			text-align: center;
		}
		.logo {
			width: 60px;
			height: 60px;
			margin: 0 auto 15px;
			background-color: rgba(255, 255, 255, 0.1);
			border-radius: 12px;
			display: flex;
			align-items: center;
			justify-content: center;
			border: 1px solid rgba(255, 255, 255, 0.2);
		}
		.logo img {
			width: 60px;
			height: 60px;
			object-fit: contain;
		}
		.header h1 {
			color: #ffffff; /* --primary-foreground */
			font-size: 22px;
			font-weight: 600;
			margin-bottom: 5px;
		}
		.header p {
			color: rgba(255, 255, 255, 0.85);
			font-size: 15px;
			font-weight: 400;
		}
		.content {
			font-family: 'Inter', -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
			padding: 35px 25px;
			background: #ffffff;
		}
		.greeting {
			font-size: 20px;
			font-weight: 600;
			color: #0f172a; /* --card-foreground */
			margin-bottom: 20px;
		}
		.message {
			font-size: 15px;
			color: #334155; /* Slightly lighter text */
			margin-bottom: 25px;
			line-height: 1.7;
		}
		.card {
			background-color: #f8fafc;
			border: 1px solid #e2e8f0;
			border-radius: 10px;
			padding: 25px;
			margin: 25px 0;
		}
		.card h3 {
			font-size: 17px;
			font-weight: 600;
			color: #0f172a; /* --card-foreground */
			margin-bottom: 15px;
		}
		.card ul {
			list-style: none;
			padding: 0;
		}
		.card li {
			font-size: 14px;
			color: #475569;
			margin-bottom: 10px;
			display: flex;
			align-items: flex-start;
		}
		.card li::before {
			content: "•";
			color: #fd7d19; /* --primary */
			font-weight: bold;
			display: inline-block;
			width: 1.2em;
			margin-left: -1.2em;
		}
		.code-section {
			background-color: #fffbeb;
			border: 1px solid #fde68a;
			border-radius: 10px;
			padding: 25px;
			margin: 25px 0;
			text-align: center;
		}
		.code-label {
			font-size: 14px;
			color: #713f12;
			font-weight: 500;
			margin-bottom: 15px;
		}
		.verification-code {
			font-size: 36px;
			font-weight: 700;
			color: #00434c; /* --secondary */
			letter-spacing: 6px;
			margin: 15px 0;
			font-family: 'Monaco', 'Menlo', monospace;
			background: #ffffff;
			padding: 15px 25px;
			border-radius: 8px;
			border: 1px solid #fed7aa;
			display: inline-block;
			box-shadow: 0 2px 4px rgba(0, 0, 0, 0.03);
		}
		.cta-section {
			text-align: center;
			margin: 30px 0;
		}
		.cta-button {
			display: inline-block;
			background-color: #fd7d19; /* --primary */
			color: #ffffff; /* --primary-foreground */
			text-decoration: none;
			padding: 14px 28px;
			border-radius: 8px;
			font-weight: 600;
			font-size: 15px;
			transition: background-color 0.2s ease;
			box-shadow: 0 2px 6px rgba(253, 125, 25, 0.2);
		}
		.cta-button:hover {
			background-color: #e66b0e; /* Darker shade of primary */
		}
		.highlight {
			background-color: #fff1e5; /* Light primary tint */
			border-left: 4px solid #fd7d19; /* --primary */
			padding: 15px 20px;
			border-radius: 0 8px 8px 0;
			margin: 20px 0;
		}
		.highlight p {
			font-size: 14px;
			color: #9a3412; /* Darker shade for contrast */
			margin: 0;
		}
		.footer {
			background-color: #f1f5f9;
			padding: 25px;
			text-align: center;
			border-top: 1px solid #e2e8f0;
		}
		.footer-logo {
			font-size: 18px;
			font-weight: 700;
			color: #0f172a; /* --card-foreground */
			margin-bottom: 10px;
		}
		.footer-text {
			font-size: 13px;
			color: #64748b;
			margin-bottom: 6px;
		}
		.footer-links {
			margin-top: 15px;
		}
		.footer-link {
			color: #00434c; /* --secondary */
			text-decoration: none;
			font-size: 13px;
			margin: 0 10px;
		}
		.footer-link:hover {
			text-decoration: underline;
		}
		.divider {
			height: 1px;
			background-color: #e2e8f0;
			margin: 25px 0;
		}
		@media (max-width: 600px) {
			.email-wrapper {
				margin: 0 10px;
				border-radius: 10px;
			}
			.header, .content, .footer {
				padding: 25px 20px;
			}
			.header h1 {
				font-size: 20px;
			}
			.greeting {
				font-size: 18px;
			}
			.verification-code {
				font-size: 28px;
				letter-spacing: 4px;
				padding: 12px 20px;
			}
			.card, .code-section {
				padding: 20px 15px;
			}
		}

`
