import { useEffect, useState, type FormEvent } from "react";
import { Link, useNavigate } from "react-router-dom";
import { ArrowRight, LoaderCircle } from "lucide-react";
import { register, sendCaptcha } from "../../api/auth";
import { errorMessage } from "../../api/client";

export default function RegisterPage() {
  const navigate = useNavigate();
  const [form, setForm] = useState({ name: "", email: "", password: "", captcha: "" });
  const [busy, setBusy] = useState(false);
  const [sendingCode, setSendingCode] = useState(false);
  const [cooldown, setCooldown] = useState(0);
  const [error, setError] = useState("");

  useEffect(() => {
    if (cooldown <= 0) return;
    const timer = window.setInterval(() => setCooldown((value) => Math.max(0, value - 1)), 1000);
    return () => window.clearInterval(timer);
  }, [cooldown]);

  const update = (field: keyof typeof form, value: string) => {
    setForm((current) => ({ ...current, [field]: value }));
  };

  const requestCode = async () => {
    if (!form.email.trim()) {
      setError("Enter your email before requesting a verification code.");
      return;
    }
    setError("");
    setSendingCode(true);
    try {
      await sendCaptcha(form.email.trim());
      setCooldown(60);
    } catch (cause) {
      setError(errorMessage(cause, "Unable to send the verification code."));
    } finally {
      setSendingCode(false);
    }
  };

  const submit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    setError("");
    setBusy(true);
    try {
      await register({ ...form, name: form.name.trim(), email: form.email.trim() });
      navigate("/login", { replace: true, state: { registered: true } });
    } catch (cause) {
      setError(errorMessage(cause, "Unable to create your account."));
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
          <h1>From first node<br />to working system.</h1>
          <p>A practical canvas for models, knowledge, tools, and control flow.</p>
        </div>
        <div className="auth-coordinate" aria-hidden="true">x 428&nbsp;&nbsp; y 216</div>
      </section>
      <section className="auth-form-panel">
        <form className="auth-form" onSubmit={submit}>
          <header>
            <h2>Create your account</h2>
            <p>Set up a workspace in a few details.</p>
          </header>
          {error ? <div className="form-alert" role="alert">{error}</div> : null}
          <label className="field"><span>Name</span><input value={form.name} onChange={(event) => update("name", event.target.value)} autoComplete="name" required autoFocus /></label>
          <label className="field"><span>Email</span><input type="email" value={form.email} onChange={(event) => update("email", event.target.value)} autoComplete="email" required /></label>
          <label className="field"><span>Password</span><input type="password" value={form.password} onChange={(event) => update("password", event.target.value)} autoComplete="new-password" minLength={6} required /><small>Use at least 6 characters.</small></label>
          <label className="field">
            <span>Verification code</span>
            <span className="field-combo">
              <input value={form.captcha} onChange={(event) => update("captcha", event.target.value)} inputMode="numeric" autoComplete="one-time-code" required />
              <button className="button button--secondary" type="button" onClick={requestCode} disabled={sendingCode || cooldown > 0}>
                {sendingCode ? "Sending…" : cooldown > 0 ? `${cooldown}s` : "Send code"}
              </button>
            </span>
          </label>
          <button className="button button--primary button--full" type="submit" disabled={busy}>
            {busy ? <LoaderCircle className="spin" size={17} aria-hidden="true" /> : null}
            {busy ? "Creating account…" : "Create account"}
            {!busy ? <ArrowRight size={17} aria-hidden="true" /> : null}
          </button>
          <p className="auth-switch">Already have an account? <Link to="/login">Sign in</Link></p>
        </form>
      </section>
    </main>
  );
}

