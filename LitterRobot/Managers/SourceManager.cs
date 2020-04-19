using System.Collections.Generic;
using System.Threading.Channels;
using System.Threading.Tasks;
using System.Threading;
using System.Linq;
using Microsoft.Extensions.Logging;
using Microsoft.Extensions.Options;
using System.Net.Http;
using LitterRobot.Models.Shared;
using TwoMQTT.Core.Managers;
using LitterRobot.Models.SourceManager;
using System;
using LitterRobot.DataAccess;
using TwoMQTT.Core;
using TwoMQTT.Core.DataAccess;

namespace LitterRobot.Managers
{
    public class SourceManager : HTTPPollingManager<SlugMapping, FetchResponse, object, Resource, Command>
    {
        public SourceManager(ILogger<SourceManager> logger, IOptions<Models.Shared.Opts> sharedOpts, IOptions<Models.SourceManager.Opts> opts, ChannelWriter<Resource> outgoing, ChannelReader<Command> incoming, IHTTPSourceDAO<SlugMapping, Command, Models.SourceManager.FetchResponse, object> sourceDAO, IHttpClientFactory httpClientFactory) :
            base(logger, outgoing, incoming, sharedOpts.Value.Resources, opts.Value.PollingInterval, sourceDAO)
        {
            this.Opts = opts.Value;
            this.SharedOpts = sharedOpts.Value;
        }

        /// <inheritdoc />
        protected override void LogSettings()
        {
            this.Logger.LogInformation(
                $"Login: {this.Opts.Login}\n" +
                $"Password: {(!string.IsNullOrEmpty(this.Opts.Password) ? "<REDACTED>" : string.Empty)}\n" +
                $"PollingInterval: {this.PollingInterval}\n" +
                $"Resources: {this.Questions.Select(x => $"{x.LRID}:{x.Slug}")}\n" +
                $""
            );
        }

        /// <inheritdoc />
        protected override Resource MapResponse(FetchResponse src) =>
            new Resource
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
                CleanCycleWaitTimeMinutes = src.CleanCycleWaitTimeMinutes,
                PanelLockActive = src.PanelLockActive == Resource.ON_ONE,
                NightLightActive = src.NightLightActive == Resource.ON_ONE,
                SleepModeActive = src.SleepModeActive == Resource.ON_ONE,
                SleepMode = src.SleepModeActive == Resource.ON_ONE ? src.SleepModeActive.Substring(1) : string.Empty,
                DFITriggered = src.IsDFITriggered == Resource.ON_ONE,
                CycleCount = src.CycleCount,
                CycleCapacity = src.CycleCapacity,
                CyclesAfterDrawerFull = src.CyclesAfterDrawerFull,
                DFICycleCount = src.DFICycleCount,
            };

        private readonly Models.SourceManager.Opts Opts;
        private readonly Models.Shared.Opts SharedOpts;
        private readonly Dictionary<string, string> StatusMapping = new Dictionary<string, string>
        {
            { "RDY", "Ready" },
            { "OFF", "Off" },
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
    }
}
