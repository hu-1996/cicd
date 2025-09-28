import { useNavigate } from "react-router-dom";
import { Tooltip, Button, message } from "antd";
import { ForwardOutlined } from "@ant-design/icons";
import { fetchRequest } from "../../../utils/fetch";
import { colors, status } from "../../../config/consts";

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
            <Button disabled={steps[index].last_status !== 'success' && steps[index].last_status !== 'failed'} type="text" size="small" className="h-[16px]" onClick={() => startNextStep(steps[index + 1].last_runner_id)}>
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
