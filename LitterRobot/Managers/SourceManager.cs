using System;
using System.Collections.Generic;
using System.Linq;
using System.Net.Http;
using System.Threading.Channels;
using Microsoft.Extensions.Logging;
using Microsoft.Extensions.Options;
using LitterRobot.Models.Shared;
using LitterRobot.Models.SourceManager;
using TwoMQTT.Core;
using TwoMQTT.Core.DataAccess;
using TwoMQTT.Core.Managers;

namespace LitterRobot.Managers
{
    /// <summary>
    /// An class representing a managed way to interact with a source.
    /// </summary>
    public class SourceManager : APIPollingManager<SlugMapping, FetchResponse, object, Resource, Command>
    {
        public SourceManager(ILogger<SourceManager> logger, IOptions<Models.Shared.Opts> sharedOpts,
            IOptions<Models.SourceManager.Opts> opts, ChannelWriter<Resource> outgoing, ChannelReader<Command> incoming,
            ISourceDAO<SlugMapping, Command, Models.SourceManager.FetchResponse, object> sourceDAO,
            IHttpClientFactory httpClientFactory) :
            base(logger, outgoing, incoming, sharedOpts.Value.Resources, opts.Value.PollingInterval, sourceDAO,
                SourceSettings(sharedOpts.Value, opts.Value))
        {
        }

        /// <inheritdoc />
        protected override Resource MapResponse(FetchResponse src)
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

        /// <summary>
        /// The translation between machine codes and human-readable statuses.
        /// </summary>
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

        private static string SourceSettings(Models.Shared.Opts sharedOpts, Models.SourceManager.Opts opts) =>
            $"Login: {opts.Login}\n" +
            $"Password: {(!string.IsNullOrEmpty(opts.Password) ? "<REDACTED>" : string.Empty)}\n" +
            $"PollingInterval: {opts.PollingInterval}\n" +
            $"Resources: {string.Join(',', sharedOpts.Resources.Select(x => $"{x.LRID}:{x.Slug}"))}\n" +
            $"";
    }
}
