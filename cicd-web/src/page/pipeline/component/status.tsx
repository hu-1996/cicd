import { useNavigate } from "react-router-dom";
import { Tooltip, Button, message } from "antd";
import { ForwardOutlined } from "@ant-design/icons";
import { fetchRequest } from "../../../utils/fetch";

const colors: any = {
  failed: "#ea5506",
  success: "#50d71e",
  running: "#0095d9",
  pending: "#afafb0",
  assigning: "#afafb0",
  partial_running: "#89c3eb",
  partial_success: "#98d98e",
};

const status: any = {
  failed: "✗",
  success: "✔️",
  running: "🚀",
  pending: "🧘‍♂️",
  assigning: "🧘‍♂️",
  partial_running: "...🚀",
  partial_success: "...✔️",
};

export default function Status(props: any) {
  const { steps } = props;

  const navigate = useNavigate();

  const startNextStep = async (nextJobRunnerId: number) => {
    await fetchRequest("/api/start_job_step/" + nextJobRunnerId, {
      method: "POST",
    });
    message.success("开始执行");
  }
  return (
    <>
      {steps?.map((step: any, index: number) => (
        <div key={step.last_runner_id} className="flex items-center">
          <Tooltip placement="leftBottom" title={step.name} key={step.last_runner_id}>
            <div
              className="h-[16px] w-[40px] bg-[#afafb0] text-center leading-[16px] cursor-pointer"
              style={{
                backgroundColor: colors[step.last_status],
              }}
              onClick={() => {
                navigate("/logs?id=" + step.last_runner_id);
              }}
            >
              {status[step.last_status]}
            </div>
          </Tooltip>
          {index < steps.length - 1 && (steps[index + 1].last_status === 'pending' || steps[index + 1].last_status === '') ? (
            <Button type="text" size="small" className="h-[16px]" onClick={() => startNextStep(steps[index + 1].last_runner_id)}>
              <ForwardOutlined className="text-[20px]" />
            </Button>
          ) : (
            <div className="mr-1"></div>
          )}
        </div>
      ))}
    </>
  );
}
