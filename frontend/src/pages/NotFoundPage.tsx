import { Link } from "react-router-dom";

export default function NotFoundPage() {
  return (
    <main className="not-found">
      <span>404</span>
      <h1>This route has no node.</h1>
      <p>The page may have moved or the address is incomplete.</p>
      <Link className="button button--primary" to="/workflows">Back to workflows</Link>
    </main>
  );
}

