import { useNavigate } from "react-router-dom";
import { Tooltip, Button, message } from "antd";
import { ForwardOutlined } from "@ant-design/icons";
import React from "react";
import { fetchRequest } from "../../../utils/fetch";
import { colors, status } from "../../../config/consts";
import { useEffect, useState } from "react";

interface StepGroup {
  stage_id: any;
  parallel: boolean;
  steps: any[];
}

export default function Status(props: any) {
  const { steps } = props;
  const [stepGroups, setStepGroups] = useState<StepGroup[]>([]);
  const navigate = useNavigate();

  useEffect(() => {
    if (!steps?.length) return;
    
    const groups: StepGroup[] = [];
    let currentGroup: StepGroup | null = null;

    steps.forEach((step: any) => {
      // Only group steps with stage_id > 0
      const shouldGroup = step.stage_id > 0;
      
      if (shouldGroup && (!currentGroup || currentGroup.stage_id !== step.stage_id)) {
        currentGroup = {
          stage_id: step.stage_id,
          parallel: step.parallel || false,
          steps: []
        };
        groups.push(currentGroup);
      } else if (!shouldGroup) {
        // Create a new group for each step without grouping
        currentGroup = {
          stage_id: step.stage_id,
          parallel: false,
          steps: [step]
        };
        groups.push(currentGroup);
        return;
      }
      
      if (currentGroup) {
        currentGroup.steps.push(step);
      }
    });

    setStepGroups(groups);
  }, [steps]);

  const startNextStep = async (nextJobRunnerId: number) => {
    await fetchRequest("/api/start_job_step/" + nextJobRunnerId, {
      method: "POST",
    });
    message.success("开始执行");
  };

  return (
    <div className="flex items-center gap-1">
      {stepGroups.map((group, groupIndex) => (
        <React.Fragment key={`group-${groupIndex}`}>
          {groupIndex > 0 &&
            (group.steps[0].last_status === "pending" || group.steps[0].last_status === "") && (
              <Button
                disabled={
                  stepGroups[groupIndex - 1].steps[stepGroups[groupIndex - 1].steps.length - 1].last_status !== "success" &&
                  stepGroups[groupIndex - 1].steps[stepGroups[groupIndex - 1].steps.length - 1].last_status !== "failed"
                }
                type="text"
                size="small"
                className="h-[16px]"
                onClick={() => startNextStep(group.steps[0].last_runner_id)}
              >
                <ForwardOutlined className="text-[20px]" />
              </Button>
            )
          }
          <div 
            key={groupIndex}
            className={`flex items-center gap-1 ${group.stage_id > 0 ? 'p-1 rounded border border-gray-400 border-dashed' : ''}`}
          >
          {group.steps.map((step, stepIndex) => (
            <div key={step.last_runner_id} className="flex items-center">
              <Tooltip
                placement="leftBottom"
                title={step.name}
              >
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
              {stepIndex < group.steps.length - 1 &&
                !group.parallel &&
                (group.steps[stepIndex + 1].last_status === "pending" ||
                  group.steps[stepIndex + 1].last_status === "") && (
                  <Button
                    disabled={
                      step.last_status !== "success" &&
                      step.last_status !== "failed"
                    }
                    type="text"
                    size="small"
                    className="h-[16px] px-0 min-w-[16px]"
                    onClick={() => startNextStep(group.steps[stepIndex + 1].last_runner_id)}
                  >
                    <ForwardOutlined className="text-[20px]" />
                  </Button>
                )}
            </div>
          ))}
          </div>
        </React.Fragment>
      ))}
    </div>
  );
}
