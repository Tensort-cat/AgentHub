import { useEffect, useRef, useState, type FormEvent } from "react";
import { Check, Circle, LoaderCircle, RotateCw, X } from "lucide-react";
import { runWorkflow, streamWorkflowRun } from "../../api/workflows";
import { errorMessage } from "../../api/client";
import { Dialog } from "../../components/ui/Dialog";
import type { WorkflowRunState } from "../../types/domain";

const stages = ["queued", "running", "success"] as const;

function RunTimeline({ state }: { state: WorkflowRunState | null }) {
  const current = state?.status ?? "idle";
  const failed = current === "failed";
  const activeIndex = failed ? 2 : stages.indexOf(current as (typeof stages)[number]);
  return (
    <ol className={`run-timeline ${failed ? "run-timeline--failed" : ""}`} aria-label="Run status">
      {stages.map((stage, index) => {
        const complete = activeIndex > index || (stage === "success" && current === "success");
        const active = activeIndex === index;
        const label = stage === "success" && failed ? "Failed" : stage[0].toUpperCase() + stage.slice(1);
        return <li className={complete ? "run-stage--complete" : active ? "run-stage--active" : ""} key={stage}>{failed && index === 2 ? <X size={14} aria-hidden="true" /> : complete ? <Check size={14} aria-hidden="true" /> : active ? <LoaderCircle className="spin" size={14} aria-hidden="true" /> : <Circle size={12} aria-hidden="true" />}<span>{label}</span></li>;
      })}
    </ol>
  );
}

export function RunPanel({ open, workflowId, onClose, onComplete }: { open: boolean; workflowId: string; onClose: () => void; onComplete: () => void }) {
  const controllerRef = useRef<AbortController | null>(null);
  const attemptRef = useRef(0);
  const [input, setInput] = useState("");
  const [state, setState] = useState<WorkflowRunState | null>(null);
  const [busy, setBusy] = useState(false);
  const [streamError, setStreamError] = useState("");

  useEffect(() => () => {
    attemptRef.current += 1;
    controllerRef.current?.abort();
  }, []);

  const connect = async (taskId: string, controller: AbortController, attempt: number) => {
    let terminal = false;
    setStreamError("");
    try {
      await streamWorkflowRun(taskId, controller.signal, (next) => {
        if (controller.signal.aborted || attemptRef.current !== attempt) return;
        setState(next);
        if (next.status === "success" || next.status === "failed") {
          terminal = true;
          setBusy(false);
          onComplete();
        }
      });
      if (!terminal && !controller.signal.aborted && attemptRef.current === attempt) {
        setBusy(false);
        setStreamError("The live connection ended before a final status was received. Reconnect to continue.");
      }
    } catch (error) {
      if (controller.signal.aborted || attemptRef.current !== attempt) return;
      setBusy(false);
      setStreamError(errorMessage(error, "The live connection ended. Reconnect to read the latest status."));
    }
  };

  const reconnect = (taskId: string) => {
    controllerRef.current?.abort();
    const attempt = ++attemptRef.current;
    const controller = new AbortController();
    controllerRef.current = controller;
    setBusy(true);
    void connect(taskId, controller, attempt);
  };

  const submit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    setBusy(true);
    setState(null);
    setStreamError("");
    controllerRef.current?.abort();
    const attempt = ++attemptRef.current;
    const controller = new AbortController();
    controllerRef.current = controller;
    try {
      const accepted = await runWorkflow(workflowId, input, controller.signal);
      if (controller.signal.aborted || attemptRef.current !== attempt) return;
      setState(accepted);
      await connect(accepted.task_id, controller, attempt);
    } catch (error) {
      if (controller.signal.aborted || attemptRef.current !== attempt) return;
      setBusy(false);
      setStreamError(errorMessage(error, "Unable to start this workflow. Check the graph and node configuration."));
    }
  };

  const close = () => {
    attemptRef.current += 1;
    controllerRef.current?.abort();
    setBusy(false);
    onClose();
  };

  const terminal = state?.status === "success" || state?.status === "failed";

  return (
    <Dialog open={open} title="Run workflow" description="Submit an input and follow the workflow-level execution state." onClose={close} size="large">
      <div className="run-panel">
        <form className="run-form" onSubmit={submit}>
          <label className="field"><span>Input <b aria-hidden="true">*</b></span><textarea name="workflow_input" autoComplete="off" value={input} onChange={(event) => setInput(event.target.value)} rows={6} placeholder="Enter the value passed to the Start node…" required disabled={busy} /></label>
          <button className="button button--primary" type="submit" disabled={busy || !input.trim()}>{busy ? <LoaderCircle className="spin" size={16} aria-hidden="true" /> : null}{busy ? "Running…" : terminal ? "Run again" : "Run workflow"}</button>
        </form>
        <section className="run-output" aria-live="polite">
          <RunTimeline state={state} />
          {streamError ? <div className="form-alert" role="alert"><span>{streamError}</span>{state?.task_id ? <button className="button button--secondary" type="button" onClick={() => reconnect(state.task_id)}><RotateCw size={15} aria-hidden="true" />Reconnect</button> : null}</div> : null}
          {!state && !streamError ? <div className="run-placeholder">Run output will appear here.</div> : null}
          {state?.task_id ? <div className="run-task"><span>Task</span><code>{state.task_id}</code></div> : null}
          {state?.status === "success" ? <div className="result-block"><header><strong>Result</strong>{state.session_id ? <span>Session {state.session_id}</span> : null}</header><pre>{state.result || "The workflow completed without a text result."}</pre></div> : null}
          {state?.status === "failed" ? <div className="result-block result-block--error"><header><strong>Run failed</strong></header><pre>{state.error || "The backend did not provide an error message."}</pre></div> : null}
        </section>
      </div>
    </Dialog>
  );
}
