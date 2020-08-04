using System;
using System.Collections.Generic;
using System.Linq;
using System.Runtime.CompilerServices;
using System.Threading;
using Microsoft.Extensions.Logging;
using Microsoft.Extensions.Options;
using LitterRobot.DataAccess;
using LitterRobot.Models.Shared;
using LitterRobot.Models.Source;
using TwoMQTT.Core;
using TwoMQTT.Core.Interfaces;

namespace LitterRobot.Liasons
{
    /// <summary>
    /// An class representing a managed way to interact with a source.
    /// </summary>
    public class SourceLiason : ISourceLiason<Resource, Command>
    {
        public SourceLiason(ILogger<SourceLiason> logger, ISourceDAO sourceDAO,
            IOptions<Models.Options.SourceOpts> opts, IOptions<Models.Options.SharedOpts> sharedOpts)
        {
            this.Logger = logger;
            this.SourceDAO = sourceDAO;
            this.Questions = sharedOpts.Value.Resources;

            this.Logger.LogInformation(
                $"Login: {opts.Value.Login}\n" +
                $"Password: {(!string.IsNullOrEmpty(opts.Value.Password) ? "<REDACTED>" : string.Empty)}\n" +
                $"PollingInterval: {opts.Value.PollingInterval}\n" +
                $"Resources: {string.Join(',', sharedOpts.Value.Resources.Select(x => $"{x.LRID}:{x.Slug}"))}\n" +
                $""
            );
        }

        /// <inheritdoc />
        public async IAsyncEnumerable<Resource?> FetchAllAsync([EnumeratorCancellation] CancellationToken cancellationToken = default)
        {
            foreach (var key in this.Questions)
            {
                this.Logger.LogDebug($"Looking up {key}");
                var result = await this.SourceDAO.FetchOneAsync(key, cancellationToken);
                var resp = result != null ? this.MapData(result) : null;
                yield return resp;
            }
        }

        /// <summary>
        /// The logger used internally.
        /// </summary>
        private readonly ILogger<SourceLiason> Logger;

        /// <summary>
        /// The dao used to interact with the source.
        /// </summary>
        private readonly ISourceDAO SourceDAO;

        /// <summary>
        /// The questions to ask the source (typically some kind of key/slug pairing).
        /// </summary>
        private readonly List<SlugMapping> Questions;

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
