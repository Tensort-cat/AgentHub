import { useState, type FormEvent } from "react";
import { Link, useNavigate } from "react-router-dom";
import { ArrowRight, Eye, EyeOff, LoaderCircle } from "lucide-react";
import { login } from "../../api/auth";
import { errorMessage } from "../../api/client";

export default function LoginPage() {
  const navigate = useNavigate();
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [showPassword, setShowPassword] = useState(false);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");

  const submit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    setError("");
    setBusy(true);
    try {
      await login(email.trim(), password);
      navigate("/workflows", { replace: true });
    } catch (cause) {
      setError(errorMessage(cause, "Unable to sign in. Check your email and password."));
    } finally {
      setBusy(false);
    }
  };

  return (
    <main className="auth-page">
      <section className="auth-context" aria-label="AgentHub introduction">
        <div className="brand-lockup brand-lockup--auth">
          <span className="brand-mark" aria-hidden="true"><i /><i /><i /></span>
          <span className="brand-name">AgentHub</span>
        </div>
        <div className="auth-context__message">
          <h1>Build the path.<br />Keep the logic visible.</h1>
          <p>Design, connect, and run AI workflows from one focused workspace.</p>
        </div>
        <div className="auth-blueprint" aria-hidden="true">
          <span className="auth-blueprint__node auth-blueprint__node--one">Input</span>
          <span className="auth-blueprint__line auth-blueprint__line--one" />
          <span className="auth-blueprint__node auth-blueprint__node--two">Reason</span>
          <span className="auth-blueprint__line auth-blueprint__line--two" />
          <span className="auth-blueprint__node auth-blueprint__node--three">Output</span>
        </div>
      </section>
      <section className="auth-form-panel">
        <form className="auth-form" onSubmit={submit}>
          <header>
            <h2>Welcome back</h2>
            <p>Sign in to continue to your workspace.</p>
          </header>
          {error ? <div className="form-alert" role="alert">{error}</div> : null}
          <label className="field">
            <span>Email</span>
            <input type="email" value={email} onChange={(event) => setEmail(event.target.value)} autoComplete="email" required autoFocus />
          </label>
          <label className="field">
            <span>Password</span>
            <span className="input-with-action">
              <input type={showPassword ? "text" : "password"} value={password} onChange={(event) => setPassword(event.target.value)} autoComplete="current-password" minLength={6} required />
              <button type="button" onClick={() => setShowPassword((value) => !value)} aria-label={showPassword ? "Hide password" : "Show password"}>
                {showPassword ? <EyeOff size={17} aria-hidden="true" /> : <Eye size={17} aria-hidden="true" />}
              </button>
            </span>
          </label>
          <button className="button button--primary button--full" type="submit" disabled={busy}>
            {busy ? <LoaderCircle className="spin" size={17} aria-hidden="true" /> : null}
            {busy ? "Signing in…" : "Sign in"}
            {!busy ? <ArrowRight size={17} aria-hidden="true" /> : null}
          </button>
          <p className="auth-switch">New to AgentHub? <Link to="/register">Create an account</Link></p>
        </form>
      </section>
    </main>
  );
}

