using System;
using System.Collections.Generic;
using System.Linq;
using System.Threading;
using System.Threading.Tasks;
using Microsoft.Extensions.Logging;
using Microsoft.Extensions.Options;
using LitterRobot.DataAccess;
using LitterRobot.Models.Options;
using LitterRobot.Models.Shared;
using LitterRobot.Models.Source;
using TwoMQTT;
using TwoMQTT.Interfaces;
using TwoMQTT.Liasons;

namespace LitterRobot.Liasons
{
    /// <summary>
    /// An class representing a managed way to interact with a source.
    /// </summary>
    public class SourceLiason : SourceLiasonBase<Resource, Command, SlugMapping, ISourceDAO, SharedOpts>, ISourceLiason<Resource, Command>
    {
        public SourceLiason(ILogger<SourceLiason> logger, ISourceDAO sourceDAO,
            IOptions<Models.Options.SourceOpts> opts, IOptions<Models.Options.SharedOpts> sharedOpts) :
            base(logger, sourceDAO, sharedOpts)
        {
            this.Logger.LogInformation(
                "Login: {login}\n" +
                "Password: {password}\n" +
                "PollingInterval: {pollingInterval}\n" +
                "Resources: {@resources}\n" +
                "",
                opts.Value.Login,
                (!string.IsNullOrEmpty(opts.Value.Password) ? "<REDACTED>" : string.Empty),
                opts.Value.PollingInterval,
                sharedOpts.Value.Resources
            );
        }

        protected override async Task<Resource?> FetchOneAsync(SlugMapping key, CancellationToken cancellationToken)
        {
            var result = await this.SourceDAO.FetchOneAsync(key, cancellationToken);
            return result switch
            {
                Response => this.MapData(result),
                _ => null,
            };
        }

        /// <summary>
        /// The translation between machine codes and human-readable statuses.
        /// </summary>
        private readonly Dictionary<string, string> StatusMapping = new Dictionary<string, string>
        {
            { "RDY", "Ready" },
            { "OFF", "Off" },
            { "OFFLINE", "Off" },
            { "P",   "Paused" },
            { "BR",  "Bonnet removed" },
            { "DFS", "Drawer full" },
            { "DF1", "Drawer full; two cycles remain" },
            { "DF2", "Drawer full; one cycle remains" },
            { "CST", "Cat sensor; timing" },
            { "CSI", "Cat sensor; interrupt" },
            { "CSF", "Cat sensor; fault" },
            { "CCP", "Cycle processing" },
            { "CCC", "Cycle complete" },
            { "EC",  "Emptying container" },
        };

        /// <summary>
        /// Map the source response to a shared response representation.
        /// </summary>
        /// <param name="src"></param>
        /// <returns></returns>
        private Resource MapData(Response src)
        {
            var sleepModeActive = src.SleepModeActive.Substring(0, 1) == Resource.ON_ONE;
            var sleepMode = string.Empty;
            if (sleepModeActive)
            {
                var smParts = src.SleepModeActive.Substring(1).Split(":");
                var smFormat = "hh:mm tt";
                var smDate = DateTime.Now - new TimeSpan(int.Parse(smParts[0]), int.Parse(smParts[1]), int.Parse(smParts[2]));
                sleepMode = $"{smDate.ToString(smFormat)} to {smDate.AddHours(8).ToString(smFormat)}";
            }

            var ccwtm = src.CleanCycleWaitTimeMinutes switch
            {
                "3" => 3,
                "7" => 7,
                "F" => 15,
                _ => 0,
            };

            return new Resource
            {
                LitterRobotId = src.LitterRobotId,
                LitterRobotSerial = src.LitterRobotSerial,
                LitterRobotNickname = src.LitterRobotNickname,
                PowerStatus = src.PowerStatus,
                UnitStatus = src.UnitStatus,
                UnitStatusText = this.StatusMapping.ContainsKey(src.UnitStatus) ?
                    this.StatusMapping[src.UnitStatus] :
                    src.UnitStatus,
                Power = src.SleepModeActive != Resource.ON_ONE && src.UnitStatus != Const.OFF,
                Cycle = src.SleepModeActive != Resource.ON_ONE && src.UnitStatus.StartsWith(Resource.CC),
                CleanCycleWaitTimeMinutes = ccwtm,
                PanelLockActive = src.PanelLockActive == Resource.ON_ONE,
                NightLightActive = src.NightLightActive == Resource.ON_ONE,
                SleepModeActive = sleepModeActive,
                SleepMode = sleepMode,
                DFITriggered = src.IsDFITriggered == Resource.ON_ONE,
                CycleCount = src.CycleCount,
                CycleCapacity = src.CycleCapacity,
                CyclesAfterDrawerFull = src.CyclesAfterDrawerFull,
                DFICycleCount = src.DFICycleCount,
            };
        }
    }
}
